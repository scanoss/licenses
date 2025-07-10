DROP TABLE IF EXISTS licenses;
CREATE TABLE licenses
(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    license_id TEXT UNIQUE NOT NULL,
    reference TEXT,
    is_deprecated_license_id BOOLEAN,
    details_url TEXT,
    reference_number INTEGER,
    name TEXT,
    see_also TEXT,
    is_osi_approved BOOLEAN,
    is_fsf_libre BOOLEAN
);


insert into licenses (id, license_id, reference,
                      is_deprecated_license_id, details_url, reference_number, name, see_also, is_osi_approved, is_fsf_libre)
values (1, 'MIT', 'https://spdx.org/licenses/MIT.html', false, 'https://spdx.org/licenses/MIT.json', 515, 'MIT License',
        '[https://opensource.org/license/mit/, http://opensource.org/licenses/MIT]', true, true),
       (2, 'CDDL-1.0', 'https://spdx.org/licenses/CDDL-1.0.html', false, 'https://spdx.org/licenses/CDDL-1.0.json', 407, 'Common Development and Distribution License 1.0',
        '[https://opensource.org/licenses/cddl1]', true, true),
       (3, 'Artistic-2.0', 'https://spdx.org/licenses/Artistic-2.0.html', false, 'https://spdx.org/licenses/Artistic-2.0.json', 217, 'Artistic License 2.0',
        '[http://www.perlfoundation.org/artistic_license_2_0, https://www.perlfoundation.org/artistic-license-20.html]', true, true),
       (4, 'BSD-3-Clause', 'https://spdx.org/licenses/BSD-3-Clause.html', false, 'https://spdx.org/licenses/BSD-3-Clause.json', 318, 'BSD 3-Clause "New" or "Revised" License',
        '[https://opensource.org/licenses/BSD-3-Clause, https://www.eclipse.org/org/documents/edl-v10.php]', true, true),
       (5, 'GPL-2.0-only', 'https://spdx.org/licenses/GPL-2.0-only.html', false, 'https://spdx.org/licenses/GPL-2.0-only.json', 456, 'GNU General Public License v2.0 only',
        '[https://www.gnu.org/licenses/old-licenses/gpl-2.0-standalone.html, https://opensource.org/licenses/GPL-2.0]', true, true),
       (6, 'Apache-2.0', 'https://spdx.org/licenses/Apache-2.0.html', false, 'https://spdx.org/licenses/Apache-2.0.json', 201, 'Apache License 2.0',
        '[https://www.apache.org/licenses/LICENSE-2.0, https://opensource.org/licenses/Apache-2.0]', true, true),
       (7, 'LGPL-2.1-only', 'https://spdx.org/licenses/LGPL-2.1-only.html', false, 'https://spdx.org/licenses/LGPL-2.1-only.json', 483, 'GNU Lesser General Public License v2.1 only',
        '[https://www.gnu.org/licenses/old-licenses/lgpl-2.1-standalone.html, https://opensource.org/licenses/LGPL-2.1]', true, true),
       (8, 'MPL-2.0', 'https://spdx.org/licenses/MPL-2.0.html', false, 'https://spdx.org/licenses/MPL-2.0.json', 531, 'Mozilla Public License 2.0',
        '[https://www.mozilla.org/en-US/MPL/2.0/, https://opensource.org/licenses/MPL-2.0]', true, true),
       (9, 'ISC', 'https://spdx.org/licenses/ISC.html', false, 'https://spdx.org/licenses/ISC.json', 466, 'ISC License',
        '[https://www.isc.org/licenses/, https://opensource.org/licenses/ISC]', true, true),
       (10, 'GPL-3.0-only', 'https://spdx.org/licenses/GPL-3.0-only.html', false, 'https://spdx.org/licenses/GPL-3.0-only.json', 457, 'GNU General Public License v3.0 only',
        '[https://www.gnu.org/licenses/gpl-3.0-standalone.html, https://opensource.org/licenses/GPL-3.0]', true, true),
       (11, 'Unlicense', 'https://spdx.org/licenses/Unlicense.html', false, 'https://spdx.org/licenses/Unlicense.json', 624, 'The Unlicense',
        '[https://unlicense.org/, https://opensource.org/licenses/unlicense]', true, null);
