// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2026 SCANOSS.COM
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 2 of the License, or
 * (at your option) any later version.
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package cache

import (
	"strings"
	"testing"
	"time"

	gomodels "github.com/scanoss/go-models/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// populateCache directly sets licenses in the cache map for unit testing.
func populateCache(c *SPDXLicenseCache, licenses []gomodels.SPDXLicenseDetail) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range licenses {
		key := strings.ToLower(licenses[i].ID)
		c.licenses[key] = &licenses[i]
	}
}

func boolPtr(b bool) *bool { return &b }

func newTestCache(interval time.Duration) *SPDXLicenseCache {
	logger := zap.NewNop().Sugar()
	return &SPDXLicenseCache{
		logger:   logger,
		licenses: make(map[string]*gomodels.SPDXLicenseDetail),
		interval: interval,
		done:     make(chan struct{}),
	}
}

func TestGetLicenseByID(t *testing.T) {
	cache := newTestCache(time.Hour)
	populateCache(cache, []gomodels.SPDXLicenseDetail{
		{ID: "MIT", Name: "MIT License", IsOsiApproved: boolPtr(true)},
		{ID: "Apache-2.0", Name: "Apache License 2.0", IsOsiApproved: boolPtr(true)},
		{ID: "GPL-3.0-only", Name: "GNU General Public License v3.0 only", IsOsiApproved: boolPtr(true)},
	})

	t.Run("exact match", func(t *testing.T) {
		detail, ok := cache.GetLicenseByID("MIT")
		assert.True(t, ok)
		assert.Equal(t, "MIT", detail.ID)
		assert.Equal(t, "MIT License", detail.Name)
	})

	t.Run("case insensitive lookup", func(t *testing.T) {
		for _, input := range []string{"mit", "Mit", "mIt", "MIT"} {
			detail, ok := cache.GetLicenseByID(input)
			assert.True(t, ok, "expected to find license for input %q", input)
			assert.Equal(t, "MIT", detail.ID)
		}
	})

	t.Run("different licenses", func(t *testing.T) {
		detail, ok := cache.GetLicenseByID("apache-2.0")
		assert.True(t, ok)
		assert.Equal(t, "Apache-2.0", detail.ID)

		detail, ok = cache.GetLicenseByID("GPL-3.0-only")
		assert.True(t, ok)
		assert.Equal(t, "GNU General Public License v3.0 only", detail.Name)
	})

	t.Run("not found", func(t *testing.T) {
		detail, ok := cache.GetLicenseByID("nonexistent")
		assert.False(t, ok)
		assert.Nil(t, detail)
	})

	t.Run("empty string", func(t *testing.T) {
		detail, ok := cache.GetLicenseByID("")
		assert.False(t, ok)
		assert.Nil(t, detail)
	})
}

func TestGetLicenseByID_EmptyCache(t *testing.T) {
	cache := newTestCache(time.Hour)

	detail, ok := cache.GetLicenseByID("MIT")
	assert.False(t, ok)
	assert.Nil(t, detail)
}

func TestRefreshInterval(t *testing.T) {
	t.Run("ticker is created on start with configured interval", func(t *testing.T) {
		cache := newTestCache(500 * time.Millisecond)
		cache.ticker = time.NewTicker(cache.interval)
		defer cache.ticker.Stop()

		// Verify ticker fires within the expected interval
		select {
		case <-cache.ticker.C:
			// Ticker fired as expected
		case <-time.After(2 * time.Second):
			t.Fatal("ticker did not fire within expected interval")
		}
	})

	t.Run("stop closes done channel and stops ticker", func(t *testing.T) {
		cache := newTestCache(100 * time.Millisecond)
		cache.ticker = time.NewTicker(cache.interval)

		cache.Stop()

		select {
		case <-cache.done:
			// done channel closed as expected
		default:
			t.Fatal("expected done channel to be closed after Stop()")
		}
	})

	t.Run("refresh loop exits on stop", func(t *testing.T) {
		cache := newTestCache(50 * time.Millisecond)
		cache.ticker = time.NewTicker(cache.interval)

		done := make(chan struct{})
		go func() {
			cache.refreshLoop()
			close(done)
		}()

		cache.Stop()

		select {
		case <-done:
			// refreshLoop exited as expected
		case <-time.After(2 * time.Second):
			t.Fatal("refreshLoop did not exit after Stop()")
		}
	})
}
