mkfifo /tmp/in_C1 /tmp/out_C1

mkfifo /tmp/in_C2 /tmp/out_C2

mkfifo /tmp/in_C3 /tmp/out_C3

./ctl -n 1 < /tmp/in_C1 > /tmp/out_C1 &

./ctl -n 2 < /tmp/in_C2 > /tmp/out_C2 &

./ctl -n 3 < /tmp/in_C3 > /tmp/out_C3 &

cat /tmp/out_C1 > /tmp/in_C2 &

cat /tmp/out_C2 > /tmp/in_C3 &

cat /tmp/out_C3 > /tmp/in_C1 &