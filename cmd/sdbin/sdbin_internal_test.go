package main

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

type testio struct {
	t     *testing.T
	seek  func(int64, int) (int64, error)
	read  func([]byte) (int, error)
	write func([]byte) (int, error)
}

func (t testio) Seek(offset int64, whence int) (int64, error) {
	t.t.Helper()
	if t.seek == nil {
		t.t.Fatal("Seek called but not defined")
	}
	return t.seek(offset, whence)
}

func (t testio) Read(b []byte) (int, error) {
	t.t.Helper()
	if t.read == nil {
		t.t.Fatal("Read called but not defined")
	}
	return t.read(b)
}

func (t testio) Write(b []byte) (int, error) {
	t.t.Helper()
	if t.read == nil {
		t.t.Fatal("Write called but not defined")
	}
	return t.write(b)
}

func TestSeekSet(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		seeker func(offset int64, whence int) (int64, error)
		offset int64
		err    error
	}{{
		func(offset int64, _ int) (int64, error) {
			return offset, nil
		},
		10,
		nil,
	}, {
		func(offset int64, _ int) (int64, error) {
			return offset - 1, nil
		},
		10,
		fmt.Errorf(`tried to seek to 10 landed at 9`),
	}, {
		func(offset int64, _ int) (int64, error) {
			return offset + 1, nil
		},
		10,
		fmt.Errorf(`tried to seek to 10 landed at 11`),
	}, {
		func(offset int64, _ int) (int64, error) {
			return 0, io.ErrUnexpectedEOF
		},
		10,
		fmt.Errorf(`unable to seek to 10 due to %w`, io.ErrUnexpectedEOF),
	}}

	for i, testCase := range testCases {
		i, testCase := i, testCase
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			err := seekSet(testio{t: t, seek: testCase.seeker}, testCase.offset)
			if testCase.err != nil {
				require.EqualError(t, err, testCase.err.Error())
			} else {
				require.NoError(t, err)
			}

		})
	}
}

func TestCheck(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		want   int
		got    int
		inErr  error
		outErr error
	}{{
		1, 1, nil, nil,
	}, {
		2, 1, nil, fmt.Errorf(`only managed 1 out of 2 bytes`),
	}, {
		1, 2, nil, fmt.Errorf(`got extra bytes, got 2 but wanted 1 bytes`),
	}, {
		1, 0, io.ErrUnexpectedEOF, io.ErrUnexpectedEOF,
	}}

	for i, testCase := range testCases {
		i, testCase := i, testCase
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			err := check(testCase.want, testCase.got, testCase.inErr)
			if testCase.outErr != nil {
				require.EqualError(t, err, testCase.outErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSanityCheck(t *testing.T) {
	goodSeek := func(offset int64, _ int) (int64, error) { return offset, nil }
	testCases := []struct {
		io  testio
		err error
	}{{
		testio{t, goodSeek, func(b []byte) (int, error) {
			return copy(b[:3], []byte{'n', 'a', 'a'}), nil
		}, nil},
		fmt.Errorf(`tried to read the stored salt but only managed 3 out of 8 bytes`),
	}, {
		testio{t, goodSeek, func(b []byte) (int, error) {
			return copy(b, []byte{'N', 'O', 'T', 'I', 'T', '6', '6', '6'}), nil
		}, nil},
		fmt.Errorf(`sanity check failed, was seeking to find "5644524143393939" but got "4e4f544954363636" are you sure this is a MiVue firmware file?`),
	}, {
		testio{t, goodSeek, func(b []byte) (int, error) {
			return copy(b, []byte{'V', 'D', 'R', 'A', 'C', '9', '9', '9'}), nil
		}, nil},
		nil,
	}}

	for i, testCase := range testCases {
		i, testCase := i, testCase
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			err := sanityCheck(testCase.io)
			if testCase.err != nil {
				require.EqualError(t, err, testCase.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
