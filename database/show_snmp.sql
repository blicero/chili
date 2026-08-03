-- /home/krylon/go/src/github.com/blicero/chili/database/show_snmp.sql
-- Time-stamp: <2026-08-03 11:21:57 krylon>
-- created on 03. 08. 2026 by Benjamin Walkenhorst
-- (c) 2026 Benjamin Walkenhorst
-- Use at your own risk!

SELECT
  d.name,
  a.atype,
  datetime(a.timestamp, 'unixepoch', 'localtime') AS tstamp,
  a.value
FROM attribute a
INNER JOIN device d ON a.dev_id = d.id
WHERE a.atype <> 3
ORDER BY a.timestamp DESC;
