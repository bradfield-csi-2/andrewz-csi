package bitmap

func deWah64(data []uint64) []uint64 {
	uLen := int(data[0])
	i := 0
	uData := make([]uint64, uLen)

	var mask63, mask62, w, maxu uint64
	w = 64
	maxu = 1 << (w - 1)
	mask63 = maxu - 1
	mask62 = mask63 >> 1
	fillMask := maxu >> 1

	space := w
	for _, c := range data[1:] {
		ctype := Literal
		if c&maxu != 0 {
			if c&fillMask == 0 {
				ctype = Fill0
			} else {
				ctype = Fill1
			}
		}
		if ctype == Literal {
			if space == w {
				uData[i] = c << 1
				space = 1
			} else if space == w-1 {
				uData[i] |= c
				i++
				space = 64
			} else {
				uData[i] |= (c >> (w - 1 - space))
				uData[i+1] |= c << (space + 1)
				space++
				i++
			}
		} else {
			fill := uint64(0)
			if ctype == Fill1 {
				fill = mask63 | maxu
			}
			fLen := c & mask62
			fbitLen := fLen * (w - 1)
			if fbitLen < space {
				uData[i] |= (fill >> (w - space)) << (space - fbitLen)
				space -= fbitLen
			} else if fbitLen < (space + w) {
				uData[i] |= fill >> (w - space)
				uData[i+1] |= fill << (fbitLen - space)
				i++
				space += w - fbitLen
			} else {
				uData[i] |= fill >> (w - space)
				nFull := int((fbitLen - space) / w)
				nExt := (fbitLen - space) % w
				i++
				for j := 0; j < nFull; j++ {
					uData[i+j] = fill
				}
				i += nFull
				uData[i] |= fill << (w - nExt)
				space = w - nExt
			}
		}
	}
	return uData
}

func wah64(data []uint64) []uint64 {

	dLen := len(data)
	newData := make([]uint64, dLen*3)

	i, j := 0, 0

	var mask63, w, maxu, left, right, rleMask2, working, rleMax1, rleMax0 uint64
	var firstFill0, firstFill1 uint64
	w = 64
	maxu = 1 << (w - 1)
	mask63 = maxu - 1
	rleMask2 = maxu | (maxu >> 1)
	rleMax0 = rleMask2 - 1
	rleMax1 = maxu | mask63

	firstFill0 = maxu + 1
	firstFill1 = rleMask2 + 1

	left = data[i]
	right = data[i+1]
	i += 2
	newData[0] = uint64(dLen)
	j++

	leftover := w
	var running uint64
	for i < dLen {
		if leftover == w {
			working = left >> 1
			leftover = 1
		} else {
			working = left << (w - 1 - leftover)
			working |= right >> (leftover + 1)
			working &= mask63
			left = right
			right = data[i]
			i++
			leftover++
		}
		newData[j] = working
		j++
	}

	var remaining []uint64

	if leftover == w {
		remaining = make([]uint64, 3)
		remaining[0] = left >> 1
		remaining[1] = (left << (w - 2)) & (right >> 2) & mask63
		remaining[2] = (right << (w - 3)) & mask63
	} else if leftover == w-1 { //else if leftover = 31 {
		remaining = make([]uint64, 3)
		remaining[0] = left & mask63
		remaining[1] = right >> 1
		remaining[2] = (right << (w - 2)) & mask63
	} else {
		remaining = make([]uint64, 2)
		working = left << (w - 1 - leftover)
		working |= right >> (leftover + 1)
		remaining[0] = working & mask63
		remaining[1] = (right << (w - 2 - leftover)) & mask63
	}

	for _, rem := range remaining {
		newData[j] = rem
		j++
	}

	k := 1

	compressed := make([]uint64, j)
	compressed[0] = newData[0]

	lastType := Literal
	for _, c := range newData[1:j] {
		ctype := Literal
		if c == 0 || c == mask63 {
			if c == 0 {
				ctype = Fill0
			} else {
				ctype = Fill1
			}
		}
		if ctype == Literal {
			if lastType != Literal {
				compressed[k] = running
				k++
			}
			compressed[k] = c
			k++
			lastType = Literal
		} else {
			if ctype == lastType {
				running++
				if running == rleMax0 || running == rleMax1 {
					compressed[k] = running
					running = 0
					k++
					lastType = Literal
				}
			} else {
				if lastType != Literal {
					compressed[k] = running
					running = 0
					k++
				}
				if c == 0 {
					running = firstFill0
				} else {
					running = firstFill1
				}
			}
			lastType = ctype
		}
	}
	if lastType != Literal {
		compressed[k] = running
	}

	return compressed[:k]
}
