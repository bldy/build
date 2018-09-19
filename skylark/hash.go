package skylark

// hashString computes the FNV hash of s.
// copied for consistency https://github.com/google/skylark/blob/f09c8ae6985f50d0fbdd2a30e4c3cfff3c4746ce/hashtable.go#L335
func hashString(s string) uint32 {
	var h uint32
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}
