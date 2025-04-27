package driver

type led struct {
	on  bool
	val byte // determines the brightness of the LED
}

func (l *led) On() {
	l.on = true
	l.val = 0x7F // bright when turned on
}

func (l *led) Off() {
	l.on = false
	l.val = 0x05 // dim when turned off
}

func (l *led) Byte() byte {
	return l.val
}

func newLed() *led {
	l := led{}
	l.Off()
	return &l
}
