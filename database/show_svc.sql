-- /home/krylon/go/src/github.com/blicero/chili/database/show_svc.sql
-- Time-stamp: <2026-08-04 12:27:12 krylon>
-- created on 03. 08. 2026 by Benjamin Walkenhorst
-- (c) 2026 Benjamin Walkenhorst
-- Use at your own risk!

SELECT
  d.name,
  d.os,
  a.atype,
  datetime(a.timestamp, 'unixepoch', 'localtime') AS tstamp,
  a.value
FROM attribute a
INNER JOIN device d ON a.dev_id = d.id
WHERE a.atype = 5
ORDER BY a.timestamp DESC;
