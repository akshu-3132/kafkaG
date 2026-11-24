# uses cmd/loadtest/main.go to benchmark
# 10B..1KB in 50B steps, then 1KB..400KB in 2KB steps (~220 sizes, ~10 min).
# Env: COUNT (default 20000), END (default 409600), STEP (default 2048),
#      SMALL_STEP (default 50). STEP=1024 takes roughly 20-25 min.
set -euo pipefail

cd "$(dirname "$0")/.."
out="${1:-results.txt}"
count="${COUNT:-20000}"
end="${END:-409600}"
step="${STEP:-2048}"
small_step="${SMALL_STEP:-50}"
work="$PWD/.bench-sweep"
BPID=""

cleanup() {
	if [ -n "$BPID" ]; then
		kill "$BPID" 2>/dev/null || true
		wait "$BPID" 2>/dev/null || true
	fi
	rm -rf "$work"
}
trap cleanup EXIT

if ss -ltn 2>/dev/null | grep -q ':9092'; then
	echo "port 9092 is in use; stop the running broker and rerun" >&2
	exit 1
fi

rm -rf "$work"
mkdir -p "$work"
go build -o "$work/broker" ./cmd/broker
go build -o "$work/loadgen" ./cmd/loadtest
echo "# $(date -Is) count=$count mode=multi n=64 fresh-broker" > "$out"

sizes=()
for ((size = 10; size < 1024; size += small_step)); do sizes+=("$size"); done
for ((size = 1024; size <= end; size += step)); do sizes+=("$size"); done
if (( ${sizes[-1]} != end )); then sizes+=("$end"); fi

for size in "${sizes[@]}"; do
	(cd "$work" && exec ./broker >/dev/null 2>&1) &
	BPID=$!
	for ((try = 0; try < 100; try++)); do
		(echo >/dev/tcp/127.0.0.1/9092) 2>/dev/null && break
		sleep 0.02
	done

	echo "size=$size" >> "$out"
	"$work/loadgen" "$size" "$count" multi 64 >> "$out"

	kill "$BPID" 2>/dev/null || true
	wait "$BPID" 2>/dev/null || true
	BPID=""
	rm -rf "$work/data"
done
