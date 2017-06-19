#!/usr/bin/awk -f

BEGIN {
    FS=":[ ]+"
    SAFE_KEYS["Parents"] = 1
    IN_BODY=0
    BODY=""
}

function json_quote(text) {
    split(text, lines, "\n")
    out=""
    for (line in lines) {
        gsub(/\\/, "\\\\", lines[line])
        gsub(/\t/, "\\t", lines[line])
        gsub(/"/, "\\\"", lines[line])
        out = out lines[line]
        if (length(lines) > 1) {
            out = out "\\n"
        }
    }
    return "\"" out "\""
}

function json_object(object) {
    i = 0
    n = length(object)
    out = "{"
    for (key in object) {
        i++
        out = out json_quote(key) ":"
        if (SAFE_KEYS[key]) {
            out = out object[key]
        } else {
            out = out json_quote(object[key])
        }
        if (i < n) { out = out "," }
    }
    out = out "}"

    return out
}

function json_array(array) {
    i = 0
    n = length(array)
    out = "["
    for (key in array) {
        i++
        out = out json_quote(array[key])
        if (i < n) { out = out "," }
    }
    out = out "]"

    return out
}

!IN_BODY && length($0) == 0 {
    IN_BODY = 1
    next
}

!IN_BODY && /^Parents:/ {
    split($2, PARENTS, " ")
    FIELDS["Parents"] = json_array(PARENTS)
    next
}

!IN_BODY && /^[-A-Za-z]+:/ {
    FIELDS[$1]=$2
    next
}

IN_BODY {
    BODY=BODY $0 "\n"
}

END {
    FIELDS["Body"] = BODY
    print json_object(FIELDS)
}
