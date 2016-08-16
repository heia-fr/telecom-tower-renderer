#!/usr/bin/env bash

t1=$(curl -s --data-binary @logo.png https://telecom-tower-renderer.appspot.com/renderImage)
t2=$(curl -s -d '{"text":"A", "fgColor":"#0000ff", "bgColor":"#000000", "fontSize":6}' https://telecom-tower-renderer.appspot.com/renderText)
t3=$(curl -s -d '{"len":11, "bgColor":"#000000"}' https://telecom-tower-renderer.appspot.com/renderSpace)
t4=$(curl -s -d '{"text":"B", "fgColor":"#003311", "bgColor":"#000000", "fontSize":8}' https://telecom-tower-renderer.appspot.com/renderText)
t5=$(echo "[$t1, $t2, $t3, $t4]" | curl -s -d @- https://telecom-tower-renderer.appspot.com/join)

echo "----- IMAGE -----"
echo $t1
echo "----- LETTER IN 6x8 FONT -----"
echo $t2
echo "----- SPACE -----"
echo $t3
echo "----- LETTER IN 8x8 FONT -----"
echo $t4
echo "----- EVERYTHING JOINED -----"
echo $t5
