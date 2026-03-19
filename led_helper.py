#!/usr/bin/env python3
from rpi_ws281x import PixelStrip, Color
import sys

LED_COUNT = 4
LED_PIN = 10          # GPIO10 = SPI MOSI
LED_BRIGHTNESS = 64
LED_FREQ_HZ = 800000
LED_DMA = 10
LED_INVERT = False
LED_CHANNEL = 0

strip = PixelStrip(LED_COUNT, LED_PIN, LED_FREQ_HZ, LED_DMA, LED_INVERT, LED_BRIGHTNESS, LED_CHANNEL, strip_type=0x00081000)  # WS2812 GRB
strip.begin()

def set_color(r, g, b):
    for i in range(LED_COUNT):
        strip.setPixelColor(i, Color(r, g, b))
    strip.show()

if __name__ == "__main__":
    if len(sys.argv) > 3:
        r, g, b = map(int, sys.argv[1:4])
        set_color(r, g, b)
    else:
        set_color(255, 0, 0)  # красный по умолчанию