package mode

type Mode uint32

const (
	M1 Mode = 1 << iota
	M2
	M3
	M4
	M5
	M6
	M7
	M8
	M9
	M10
	M11
	M12
	M13
	M14
	M15
	M16
	M17
	M18
	M19
	M20
)

func (m Mode) Eq(m2 Mode) bool {
	return m&m2 != 0
}
