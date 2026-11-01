package main

import (
	"context"
	"fmt"
	"io"
)

// warmupAssessment distinguishes an interrupted transfer with useful bytes
// from a complete warmup. An interrupted transfer is never marked clean.
func warmupAssessment(want, got int64, readErr error) (complete bool, usable bool, reason string) {
	if want <= 0 {
		return true, true, ""
	}
	if got >= want {
		return true, true, ""
	}
	if got < 100000 {
		return false, false, fmt.Sprintf("预热收到 %d/%d 字节（不足 100 KB）：%v", got, want, readErr)
	}
	return false, true, fmt.Sprintf("预热传输未完成（%d/%d 字节）：%v", got, want, readErr)
}

func partialBenchmark(opened, requested int, prematurelyClosed bool) bool {
	return opened < requested || prematurelyClosed
}

func readAtMost(ctx context.Context, body io.Reader, want int64) (int64, error) {
	buf := make([]byte, 128*1024)
	var total int64
	for total < want {
		select {
		case <-ctx.Done():
			return total, ctx.Err()
		default:
		}
		next := int64(len(buf))
		if remain := want - total; remain < next {
			next = remain
		}
		n, err := body.Read(buf[:int(next)])
		if n > 0 {
			total += int64(n)
		}
		if err != nil {
			if err == io.EOF {
				if total < want {
					return total, io.ErrUnexpectedEOF
				}
				return total, nil
			}
			return total, err
		}
	}
	return total, nil
}
