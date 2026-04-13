#!/bin/bash
# Automated tests for ipk-L4-scan
# Run with: make test

BINARY="./ipk-L4-scan"
PASSED=0
FAILED=0

# ============================================================
# CONFIGURATION - Edit these values based on your environment
# ============================================================
# Network interface to use for remote scanning
# FITVPN: typically "enp0s2" or "eth0"
# Kolejnet: typically "eth0" or "wlan0"
REMOTE_IFACE="eth0"

# Target IP of the VUT reference scan server
# FITVPN:   100.65.80.146
# Kolejnet: 147.229.192.165
REMOTE_TARGET="100.65.80.146"

# Set to 1 to run remote scan tests, 0 to skip
RUN_REMOTE=0
# ============================================================

# Colors for output
GREEN="\033[0;32m"
RED="\033[0;31m"
NC="\033[0m"

pass() {
    echo -e "  ${GREEN}[PASS]${NC} $1"
    PASSED=$((PASSED + 1))
}

fail() {
    echo -e "  ${RED}[FAIL]${NC} $1"
    FAILED=$((FAILED + 1))
}

echo "=============================="
echo " ipk-L4-scan Automated Tests"
echo "=============================="

# --------------------------------------------------
# 1. Binary exists
# --------------------------------------------------
echo ""
echo "--- Build ---"
if [ -x "$BINARY" ]; then
    pass "Binary exists and is executable"
else
    fail "Binary not found or not executable"
    echo "Run 'make' first."
    exit 1
fi

# --------------------------------------------------
# 2. Help output (-h)
# --------------------------------------------------
echo ""
echo "--- Help output ---"

OUTPUT=$($BINARY -h 2>/dev/null)
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
    pass "-h returns exit code 0"
else
    fail "-h returns exit code $EXIT_CODE (expected 0)"
fi

if echo "$OUTPUT" | grep -qi "usage"; then
    pass "-h prints usage information to stdout"
else
    fail "-h does not print usage information"
fi

# --help variant
OUTPUT=$($BINARY --help 2>/dev/null)
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
    pass "--help returns exit code 0"
else
    fail "--help returns exit code $EXIT_CODE (expected 0)"
fi

# --------------------------------------------------
# 3. Interface listing (-i alone)
# --------------------------------------------------
echo ""
echo "--- Interface listing ---"

echo "       Command: $BINARY -i"
OUTPUT=$($BINARY -i 2>/dev/null)
EXIT_CODE=$?
echo "       Output:"
echo "$OUTPUT" | sed 's/^/         /'
echo ""

if [ $EXIT_CODE -eq 0 ]; then
    pass "-i returns exit code 0"
else
    fail "-i returns exit code $EXIT_CODE (expected 0)"
fi

if echo "$OUTPUT" | grep -q "lo"; then
    pass "-i lists loopback interface"
else
    fail "-i does not list loopback interface"
fi

# --------------------------------------------------
# 4. Missing arguments
# --------------------------------------------------
echo ""
echo "--- Argument validation ---"

# No arguments at all
$BINARY 2>/dev/null
if [ $? -ne 0 ]; then
    pass "No arguments returns non-zero exit code"
else
    fail "No arguments should return non-zero exit code"
fi

# Missing host
$BINARY -i lo -t 80 2>/dev/null
if [ $? -ne 0 ]; then
    pass "Missing HOST returns non-zero exit code"
else
    fail "Missing HOST should return non-zero exit code"
fi

# Missing interface
$BINARY -t 80 localhost 2>/dev/null
if [ $? -ne 0 ]; then
    pass "Missing -i returns non-zero exit code"
else
    fail "Missing -i should return non-zero exit code"
fi

# Missing ports
$BINARY -i lo localhost 2>/dev/null
if [ $? -ne 0 ]; then
    pass "Missing ports returns non-zero exit code"
else
    fail "Missing ports should return non-zero exit code"
fi

# Unknown argument
$BINARY --xyz 2>/dev/null
if [ $? -ne 0 ]; then
    pass "Unknown argument returns non-zero exit code"
else
    fail "Unknown argument should return non-zero exit code"
fi

# --------------------------------------------------
# 5. Error messages go to stderr
# --------------------------------------------------
echo ""
echo "--- Output channels ---"

STDOUT_OUTPUT=$($BINARY 2>/dev/null)
if [ -z "$STDOUT_OUTPUT" ]; then
    pass "Error messages are not printed to stdout"
else
    fail "Error messages leak to stdout"
fi

STDERR_OUTPUT=$($BINARY 2>&1 1>/dev/null)
if [ -n "$STDERR_OUTPUT" ]; then
    pass "Error messages are printed to stderr"
else
    fail "No error message on stderr for missing arguments"
fi

# --------------------------------------------------
# 6. TCP scan on localhost (requires sudo)
# --------------------------------------------------
echo ""
echo "--- Network scan (requires sudo) ---"

if [ "$(id -u)" -eq 0 ]; then

    # TCP OPEN: start a listener, scan it
    echo ""
    echo "  [TCP open]"
    nc -l -p 44441 &>/dev/null &
    NC_PID=$!
    sleep 0.3
    echo "       Command: $BINARY -i lo -t 44441 -w 500 127.0.0.1"
    OUTPUT=$(timeout 5 $BINARY -i lo -t 44441 -w 500 127.0.0.1 2>/dev/null)
    echo "       Output:  $OUTPUT"
    kill $NC_PID 2>/dev/null; wait $NC_PID 2>/dev/null

    if echo "$OUTPUT" | grep -qE "^127\.0\.0\.1 44441 tcp open$"; then
        pass "TCP open detected correctly"
    else
        fail "TCP open not detected: '$OUTPUT'"
    fi

    # TCP CLOSED: scan a port with no listener
    echo ""
    echo "  [TCP closed]"
    echo "       Command: $BINARY -i lo -t 44442 -w 500 127.0.0.1"
    OUTPUT=$(timeout 5 $BINARY -i lo -t 44442 -w 500 127.0.0.1 2>/dev/null)
    echo "       Output:  $OUTPUT"

    if echo "$OUTPUT" | grep -qE "^127\.0\.0\.1 44442 tcp closed$"; then
        pass "TCP closed detected correctly"
    else
        fail "TCP closed not detected: '$OUTPUT'"
    fi

    # TCP FILTERED: add iptables DROP rule
    echo ""
    echo "  [TCP filtered]"
    if command -v iptables &>/dev/null && iptables -A INPUT -p tcp --dport 44443 -j DROP 2>/dev/null; then
        echo "       Command: $BINARY -i lo -t 44443 -w 500 127.0.0.1"
        OUTPUT=$(timeout 10 $BINARY -i lo -t 44443 -w 500 127.0.0.1 2>/dev/null)
        echo "       Output:  $OUTPUT"
        iptables -D INPUT -p tcp --dport 44443 -j DROP 2>/dev/null

        if echo "$OUTPUT" | grep -qE "^127\.0\.0\.1 44443 tcp filtered$"; then
            pass "TCP filtered detected correctly"
        else
            fail "TCP filtered not detected: '$OUTPUT'"
        fi
    else
        echo "  [SKIP] TCP filtered test skipped (iptables not available)"
    fi

    # UDP CLOSED: no listener, kernel sends ICMP port unreachable
    echo ""
    echo "  [UDP closed]"
    echo "       Command: $BINARY -i lo -u 44444 -w 500 127.0.0.1"
    OUTPUT=$(timeout 5 $BINARY -i lo -u 44444 -w 500 127.0.0.1 2>/dev/null)
    echo "       Output:  $OUTPUT"

    if echo "$OUTPUT" | grep -qE "^127\.0\.0\.1 44444 udp closed$"; then
        pass "UDP closed detected correctly"
    else
        fail "UDP closed not detected: '$OUTPUT'"
    fi

    # UDP OPEN: start a UDP listener, no ICMP sent
    echo ""
    echo "  [UDP open]"
    nc -u -l -p 44445 &>/dev/null &
    NC_PID=$!
    sleep 0.3
    echo "       Command: $BINARY -i lo -u 44445 -w 500 127.0.0.1"
    OUTPUT=$(timeout 5 $BINARY -i lo -u 44445 -w 500 127.0.0.1 2>/dev/null)
    echo "       Output:  $OUTPUT"
    kill $NC_PID 2>/dev/null; wait $NC_PID 2>/dev/null

    if echo "$OUTPUT" | grep -qE "^127\.0\.0\.1 44445 udp open$"; then
        pass "UDP open detected correctly"
    else
        fail "UDP open not detected: '$OUTPUT'"
    fi

else
    echo "  [SKIP] Network scan tests skipped (not running as root)"
fi

# --------------------------------------------------
# 7. IPv6 scan (localhost ::1, requires sudo)
# --------------------------------------------------
echo ""
echo "--- IPv6 scan (::1, requires sudo) ---"

if [ "$(id -u)" -eq 0 ]; then

    # TCP open IPv6
    echo ""
    echo "  [IPv6 TCP open]"
    nc -6 -l 44451 &>/dev/null &
    NC_PID=$!
    sleep 0.3
    echo "       Command: $BINARY -i lo -t 44451 -w 500 ::1"
    OUTPUT=$(timeout 5 $BINARY -i lo -t 44451 -w 500 ::1 2>/dev/null)
    echo "       Output:  $OUTPUT"
    kill $NC_PID 2>/dev/null; wait $NC_PID 2>/dev/null

    if echo "$OUTPUT" | grep -qE "^::1 44451 tcp open$"; then
        pass "IPv6 TCP open detected correctly"
    else
        fail "IPv6 TCP open not detected: '$OUTPUT'"
    fi

    # TCP closed IPv6
    echo ""
    echo "  [IPv6 TCP closed]"
    echo "       Command: $BINARY -i lo -t 44452 -w 500 ::1"
    OUTPUT=$(timeout 5 $BINARY -i lo -t 44452 -w 500 ::1 2>/dev/null)
    echo "       Output:  $OUTPUT"

    if echo "$OUTPUT" | grep -qE "^::1 44452 tcp closed$"; then
        pass "IPv6 TCP closed detected correctly"
    else
        fail "IPv6 TCP closed not detected: '$OUTPUT'"
    fi

    # UDP closed IPv6
    echo ""
    echo "  [IPv6 UDP closed]"
    echo "       Command: $BINARY -i lo -u 44453 -w 500 ::1"
    OUTPUT=$(timeout 5 $BINARY -i lo -u 44453 -w 500 ::1 2>/dev/null)
    echo "       Output:  $OUTPUT"

    if echo "$OUTPUT" | grep -qE "^::1 44453 udp closed$"; then
        pass "IPv6 UDP closed detected correctly"
    else
        fail "IPv6 UDP closed not detected: '$OUTPUT'"
    fi

    # UDP open IPv6
    echo ""
    echo "  [IPv6 UDP open]"
    nc -6 -u -l 44454 &>/dev/null &
    NC_PID=$!
    sleep 0.3
    echo "       Command: $BINARY -i lo -u 44454 -w 500 ::1"
    OUTPUT=$(timeout 5 $BINARY -i lo -u 44454 -w 500 ::1 2>/dev/null)
    echo "       Output:  $OUTPUT"
    kill $NC_PID 2>/dev/null; wait $NC_PID 2>/dev/null

    if echo "$OUTPUT" | grep -qE "^::1 44454 udp open$"; then
        pass "IPv6 UDP open detected correctly"
    else
        fail "IPv6 UDP open not detected: '$OUTPUT'"
    fi

else
    echo "  [SKIP] IPv6 scan tests skipped (not running as root)"
fi

# --------------------------------------------------
# 8. Remote scan 
# --------------------------------------------------
echo ""
echo "--- Remote scan ($REMOTE_TARGET via $REMOTE_IFACE) ---"

if [ "$(id -u)" -ne 0 ]; then
    echo "  [SKIP] Remote scan tests skipped (not running as root)"
elif [ "$RUN_REMOTE" -ne 1 ]; then
    echo "  [SKIP] Remote scan tests skipped (RUN_REMOTE=0)"
else
    # TCP OPEN (guaranteed ports: 9001, 9002, 9003)
    echo ""
    echo "  [Remote TCP open]"
    for PORT in 9001 9002 9003; do
        echo "       Command: $BINARY -i $REMOTE_IFACE -t $PORT $REMOTE_TARGET"
        OUTPUT=$(timeout 10 $BINARY -i $REMOTE_IFACE -t $PORT $REMOTE_TARGET 2>/dev/null)
        echo "       Output:  $OUTPUT"
        if echo "$OUTPUT" | grep -qE "^$REMOTE_TARGET $PORT tcp open$"; then
            pass "TCP port $PORT open"
        else
            fail "TCP port $PORT expected open: '$OUTPUT'"
        fi
    done

    # TCP CLOSED (guaranteed ports: 9011, 9012, 9013)
    echo ""
    echo "  [Remote TCP closed]"
    for PORT in 9011 9012 9013; do
        echo "       Command: $BINARY -i $REMOTE_IFACE -t $PORT $REMOTE_TARGET"
        OUTPUT=$(timeout 10 $BINARY -i $REMOTE_IFACE -t $PORT $REMOTE_TARGET 2>/dev/null)
        echo "       Output:  $OUTPUT"
        if echo "$OUTPUT" | grep -qE "^$REMOTE_TARGET $PORT tcp closed$"; then
            pass "TCP port $PORT closed"
        else
            fail "TCP port $PORT expected closed: '$OUTPUT'"
        fi
    done

    # TCP FILTERED (guaranteed ports: 9021, 9022, 9023)
    echo ""
    echo "  [Remote TCP filtered]"
    for PORT in 9021 9022 9023; do
        echo "       Command: $BINARY -i $REMOTE_IFACE -t $PORT $REMOTE_TARGET"
        OUTPUT=$(timeout 10 $BINARY -i $REMOTE_IFACE -t $PORT $REMOTE_TARGET 2>/dev/null)
        echo "       Output:  $OUTPUT"
        if echo "$OUTPUT" | grep -qE "^$REMOTE_TARGET $PORT tcp filtered$"; then
            pass "TCP port $PORT filtered"
        else
            fail "TCP port $PORT expected filtered: '$OUTPUT'"
        fi
    done

    # UDP OPEN (guaranteed ports: 9031, 9032, 9033)
    echo ""
    echo "  [Remote UDP open]"
    for PORT in 9031 9032 9033; do
        echo "       Command: $BINARY -i $REMOTE_IFACE -u $PORT $REMOTE_TARGET"
        OUTPUT=$(timeout 10 $BINARY -i $REMOTE_IFACE -u $PORT $REMOTE_TARGET 2>/dev/null)
        echo "       Output:  $OUTPUT"
        if echo "$OUTPUT" | grep -qE "^$REMOTE_TARGET $PORT udp open$"; then
            pass "UDP port $PORT open"
        else
            fail "UDP port $PORT expected open: '$OUTPUT'"
        fi
    done

    # UDP CLOSED (guaranteed ports: 9041, 9042, 9043)
    echo ""
    echo "  [Remote UDP closed]"
    for PORT in 9041 9042 9043; do
        echo "       Command: $BINARY -i $REMOTE_IFACE -u $PORT $REMOTE_TARGET"
        OUTPUT=$(timeout 10 $BINARY -i $REMOTE_IFACE -u $PORT $REMOTE_TARGET 2>/dev/null)
        echo "       Output:  $OUTPUT"
        if echo "$OUTPUT" | grep -qE "^$REMOTE_TARGET $PORT udp closed$"; then
            pass "UDP port $PORT closed"
        else
            fail "UDP port $PORT expected closed: '$OUTPUT'"
        fi
    done

    # UDP FILTERED (guaranteed ports: 9051, 9052, 9053)
    # NOTE: Assignment specs - shows as open.
    echo ""
    echo "  [Remote UDP filtered (expected: open per spec)]"
    for PORT in 9051 9052 9053; do
        echo "       Command: $BINARY -i $REMOTE_IFACE -u $PORT $REMOTE_TARGET"
        OUTPUT=$(timeout 10 $BINARY -i $REMOTE_IFACE -u $PORT $REMOTE_TARGET 2>/dev/null)
        echo "       Output:  $OUTPUT"
        if echo "$OUTPUT" | grep -qE "^$REMOTE_TARGET $PORT udp open$"; then
            pass "UDP port $PORT filtered (shown as open per spec)"
        else
            fail "UDP port $PORT expected open (filtered): '$OUTPUT'"
        fi
    done

    # Random port range test
    echo ""
    echo "  [Remote port range scan]"
    echo "       Command: $BINARY -i $REMOTE_IFACE -t 9001,9012,9023 -u 9032,9043 $REMOTE_TARGET"
    OUTPUT=$(timeout 30 $BINARY -i $REMOTE_IFACE -t 9001,9012,9023 -u 9032,9043 $REMOTE_TARGET 2>/dev/null)
    echo "       Output:"
    echo "$OUTPUT" | sed 's/^/         /'
    LINES=$(echo "$OUTPUT" | wc -l)
    if [ "$LINES" -eq 5 ]; then
        pass "Combined scan returned 5 lines (3 TCP + 2 UDP)"
    else
        fail "Combined scan returned $LINES lines (expected 5)"
    fi
fi

# --------------------------------------------------
# Summary
# --------------------------------------------------
echo ""
echo "=============================="
echo " Results: $PASSED passed, $FAILED failed"
echo "=============================="

if [ $FAILED -ne 0 ]; then
    exit 1
fi
exit 0
