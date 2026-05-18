#!/usr/bin/env bash
# Big ASCII clock showing current HH:MM, plus a centered date line beneath.

t=$(date "+%H:%M")
h1=${t:0:1}; h2=${t:1:1}
m1=${t:3:1}; m2=${t:4:1}

d0=("██████" "██  ██" "██  ██" "██  ██" "██████")
d1=("    ██" "    ██" "    ██" "    ██" "    ██")
d2=("██████" "    ██" "██████" "██    " "██████")
d3=("██████" "    ██" " █████" "    ██" "██████")
d4=("██  ██" "██  ██" "██████" "    ██" "    ██")
d5=("██████" "██    " "██████" "    ██" "██████")
d6=("██████" "██    " "██████" "██  ██" "██████")
d7=("██████" "    ██" "   ██ " "  ██  " " ██   ")
d8=("██████" "██  ██" "██████" "██  ██" "██████")
d9=("██████" "██  ██" "██████" "    ██" "██████")
col=("      " "  ██  " "      " "  ██  " "      ")

row_for() {
  case "$1" in
    0) printf '%s' "${d0[$2]}" ;;
    1) printf '%s' "${d1[$2]}" ;;
    2) printf '%s' "${d2[$2]}" ;;
    3) printf '%s' "${d3[$2]}" ;;
    4) printf '%s' "${d4[$2]}" ;;
    5) printf '%s' "${d5[$2]}" ;;
    6) printf '%s' "${d6[$2]}" ;;
    7) printf '%s' "${d7[$2]}" ;;
    8) printf '%s' "${d8[$2]}" ;;
    9) printf '%s' "${d9[$2]}" ;;
    :) printf '%s' "${col[$2]}" ;;
  esac
}

echo
for row in 0 1 2 3 4; do
  printf "    "
  row_for "$h1" "$row"; printf "  "
  row_for "$h2" "$row"; printf "  "
  row_for ":"    "$row"; printf "  "
  row_for "$m1" "$row"; printf "  "
  row_for "$m2" "$row"
  echo
done
echo
date "+    %A · %B %d, %Y"
