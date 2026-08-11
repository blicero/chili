// Time-stamp: <2026-08-11 16:39:04 krylon>
// -*- mode: javascript; coding: utf-8; -*-
// Copyright 2015-2020 Benjamin Walkenhorst <krylon@gmx.net>
//
// This file has grown quite a bit larger than I had anticipated.
// It is not a /big/ problem right now, but in the long run, I will have to
// break this thing up into several smaller files.

'use strict';

const whitespace_pat = /^\s*$/

function defined(x) {
    return undefined !== x && null !== x
}

function fmtDateNumber(n) {
    return (n < 10 ? '0' : '') + n.toString()
} // function fmtDateNumber(n)

function timeStampString(t) {
    if ((typeof t) === 'string') {
        return t
    }

    const year = t.getYear() + 1900
    const month = fmtDateNumber(t.getMonth() + 1)
    const day = fmtDateNumber(t.getDate())
    const hour = fmtDateNumber(t.getHours())
    const minute = fmtDateNumber(t.getMinutes())
    const second = fmtDateNumber(t.getSeconds())

    const s =
          year + '-' + month + '-' + day +
          ' ' + hour + ':' + minute + ':' + second
    return s
} // function timeStampString(t)

function fmtDuration(seconds) {
    let minutes = 0
    let hours = 0

    while (seconds > 3599) {
        hours++
        seconds -= 3600
    }

    while (seconds > 59) {
        minutes++
        seconds -= 60
    }

    if (hours > 0) {
        return `${hours}h${minutes}m${seconds}s`
    } else if (minutes > 0) {
        return `${minutes}m${seconds}s`
    } else {
        return `${seconds}s`
    }
} // function fmtDuration(seconds)

function beaconLoop() {
    try {
        if (settings.beacon.active) {
            const req = $.get('/ajax/beacon',
                              {},
                              function (response) {
                                  let status = ''

                                  if (response.status) {
                                      status = 
                                          response.message +
                                          ' running on ' +
                                          response.hostname +
                                          ' is alive at ' +
                                          response.timestamp
                                  } else {
                                      status = 'Server is not responding'
                                  }

                                  const beaconDiv = $('#beacon')[0]

                                  if (defined(beaconDiv)) {
                                      beaconDiv.innerHTML = status
                                      beaconDiv.classList.remove('error')
                                  } else {
                                      console.log('Beacon field was not found')
                                  }
                              },
                              'json'
                             ).fail(function () {
                                 const beaconDiv = $('#beacon')[0]
                                 beaconDiv.innerHTML = 'Server is not responding'
                                 beaconDiv.classList.add('error')
                                 // logMsg("ERROR", "Server is not responding");
                             })
        }
    } finally {
        window.setTimeout(beaconLoop, settings.beacon.interval)
    }
} // function beaconLoop()

function beaconToggle() {
    settings.beacon.active = !settings.beacon.active
    saveSetting('beacon', 'active', settings.beacon.active)

    if (!settings.beacon.active) {
        const beaconDiv = $('#beacon')[0]
        beaconDiv.innerHTML = 'Beacon is suspended'
        beaconDiv.classList.remove('error')
    }
} // function beaconToggle()

/*
  The ‘content’ attribute of Window objects is deprecated.  Please use ‘window.top’ instead. interact.js:125:8
  Ignoring get or set of property that has [LenientThis] because the “this” object is incorrect. interact.js:125:8

*/

function db_maintenance() {
    const maintURL = '/ajax/db_maint'

    const req = $.get(
        maintURL,
        {},
        function (res) {
            if (!res.Status) {
                console.log(res.Message)
                msg_add('ERROR', res.Message)
            } else {
                const msg = 'Database Maintenance performed without errors'
                console.log(msg)
                msg_add('INFO', msg)
            }
        },
        'json'
    ).fail(function () {
        const msg = 'Error performing DB maintenance'
        console.log(msg)
        msg_add('ERROR', msg)
    })
} // function db_maintenance()

function run_probe() {
    const runURL = "/ajax/run-probe"
    const req = $.get(
        runURL,
        {},
        function (res) {
            if (!res.Status) {
                console.log(res.Message)
                msg_add('ERROR', res.Message)
            } else {
                const msg = res.Message
                console.log(msg)
                msg_add('INFO', msg)
            }
        },
        'json'
    ).fail(function () {
        const msg = res.Message
        console.log(msg)
        msg_add('ERROR', msg)
    })
} // function run_probe()

function scale_images() {
    const selector = '.news-body > img'
    const maxHeight = 300
    const maxWidth = 300

    $(selector).each(function () {
        const img = $(this)[0]
        if (img.width > maxWidth || img.height > maxHeight) {
            const size = shrink_img(img.width, img.height, maxWidth, maxHeight)

            img.width = size.width
            img.height = size.height
        }
    })
} // function scale_images()

// Found here: https://stackoverflow.com/questions/3971841/how-to-resize-images-proportionally-keeping-the-aspect-ratio#14731922
function shrink_img(srcWidth, srcHeight, maxWidth, maxHeight) {
    const ratio = Math.min(maxWidth / srcWidth, maxHeight / srcHeight)

    return { width: srcWidth * ratio, height: srcHeight * ratio }
} // function shrink_img(srcWidth, srcHeight, maxWidth, maxHeight)

const max_msg_cnt = 5

function msg_clear() {
    $('#msg_tbl')[0].innerHTML = ''
} // function msg_clear()

function msg_add(msg, level=1) {
    const row = `<tr><td>${new Date()}</td><td>${level}</td><td>${msg}</td><td></td></tr>`
    const msg_tbl = $('#msg_tbl')[0]

    const rows = $('#msg_tbl tr')
    let i = 0
    let cnt = rows.length
    while (cnt >= max_msg_cnt) {
        rows[i].remove()
        i++
        cnt--
    }

    msg_tbl.innerHTML += row
} // function msg_add(msg)

function fmtNumber(n, kind = "") {
    if (kind in formatters) {
        return formatters[kind](n)
    } else {
        return fmtDefault(n)
    }
} // function fmtNumber(n, kind = "")

function fmtDefault(n) {
    return n.toPrecision(3).toString()
} // function fmtDefault(n)

function fmtBytes(n) {
    const units = ["KB", "MB", "GB", "TB", "PB"]
    let idx = 0
    while (n >= 1024) {
        n /= 1024
        idx++
    }

    return `${n.toPrecision(3)} ${units[idx]}`
} // function fmtBytes(n)

const formatters = {
    "sysload": fmtNumber,
    "disk": fmtBytes,
}

function net_form_reset() {
    $(".net_form").each(() => {
        $(this).text("")
    })
} // function net_form_reset()

function net_form_submit() {
    const url = "/ajax/net/create"
    const name = $("#net_name")[0].value
    const addr = $("#net_addr")[0].value

    $.post(
        url,
        { "name": name, "addr": addr },
        (res) => {
            if (res.status) {
                const tbody = $("#net_list")[0]
                const net = res.net
                let row = `<tr>
<td>${net.id}</td>
<td><a href="/network/${net.id}">${net.name}</a></td>
<td>${net.addr}</td>
<td></td>
</tr>`
                tbody.innerHTML += row
                net_form_reset()
            } else {
                const msg = `Failed to add network: ${res.message}`
                console.error(msg)
                msg_add(msg, 2)
            }
        },
        'json').fail((xhr, txt, err) => {
            console.error(txt, err)
        })
}
