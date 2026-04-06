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
	"context"
	"strings"
	"sync"
	"time"

	gomodels "github.com/scanoss/go-models/pkg/models"
	"github.com/scanoss/go-models/pkg/scanoss"
	"go.uber.org/zap"
)

type SPDXLicenseCacheInterface interface {
	GetLicenseByID(spdxID string) (*gomodels.SPDXLicenseDetail, bool)
	Start(ctx context.Context) error
	Stop()
}

type SPDXLicenseCache struct {
	mu       sync.RWMutex
	licenses map[string]*gomodels.SPDXLicenseDetail
	sc       *scanoss.Client
	logger   *zap.SugaredLogger
	ticker   *time.Ticker
	done     chan struct{}
	interval time.Duration
}

func NewSPDXLicenseCache(sc *scanoss.Client, logger *zap.SugaredLogger, interval time.Duration) *SPDXLicenseCache {
	return &SPDXLicenseCache{
		sc:       sc,
		logger:   logger,
		licenses: make(map[string]*gomodels.SPDXLicenseDetail),
		interval: interval,
		done:     make(chan struct{}),
	}
}

// Start performs the initial load and starts the background refresh goroutine.
func (c *SPDXLicenseCache) Start(ctx context.Context) error {
	if err := c.loadFromDB(ctx); err != nil {
		return err
	}
	c.ticker = time.NewTicker(c.interval)
	go c.refreshLoop()
	return nil
}

// Stop stops the background refresh goroutine.
func (c *SPDXLicenseCache) Stop() {
	close(c.done)
	if c.ticker != nil {
		c.ticker.Stop()
	}
}

// GetLicenseByID returns the cached SPDX license detail for the given ID.
func (c *SPDXLicenseCache) GetLicenseByID(spdxID string) (*gomodels.SPDXLicenseDetail, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	detail, ok := c.licenses[strings.ToLower(spdxID)]
	return detail, ok
}

func (c *SPDXLicenseCache) loadFromDB(ctx context.Context) error {
	details, err := c.sc.Models.Licenses.GetAllSPDXLicensesDetails(ctx)
	if err != nil {
		return err
	}
	newMap := make(map[string]*gomodels.SPDXLicenseDetail, len(details))
	for i := range details {
		key := strings.ToLower(details[i].ID)
		newMap[key] = &details[i]
	}
	c.mu.Lock()
	c.licenses = newMap
	c.mu.Unlock()
	c.logger.Infof("SPDX license cache loaded: %d licenses", len(newMap))
	return nil
}

func (c *SPDXLicenseCache) refreshLoop() {
	for {
		select {
		case <-c.ticker.C:
			if err := c.loadFromDB(context.Background()); err != nil {
				c.logger.Errorf("Failed to refresh SPDX license cache: %v", err)
			}
		case <-c.done:
			return
		}
	}
}
