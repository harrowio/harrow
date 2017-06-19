package posix

type strMapFn func(rune) rune

// SafeStr maps a rune to the POSIX Portable Filename Character set.  Any rune
// outside of the range defined by POSIX is rendered as '_'.
//
// http://pubs.opengroup.org/onlinepubs/9699919799//basedefs/V1_chap03.html#tag_03_278
var SafeStr strMapFn = func(r rune) rune {
	switch {
	case r >= 'A' && r <= 'Z':
		fallthrough
	case r >= 'a' && r <= 'z':
		fallthrough
	case r >= '0' && r <= '9':
		fallthrough
	case r == '.':
		fallthrough
	case r == '_':
		fallthrough
	case r == '-':
		return r
	default:
		return '_'
	}
}
