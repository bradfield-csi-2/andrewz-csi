#!/bin/bash
BIRTHDATE="Jan 1, 2000"
Presents=10

BIRTHDAY=$(date -j -f "%b %d, %Y" "${BIRTHDATE}" "+%A")

echo $BIRTHDATE
echo $Presents
echo $BIRTHDAY
