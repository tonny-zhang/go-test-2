package main

import (
	"testing"

	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

func Test_on(t *testing.T) {
	event := Event{}
	eventSub := EventSub{}
	isGetTest1 := false
	isGetTest2 := false
	isGetTest3 := false
	event.on("test", func(argv ...interface{}) {
		assert.Assert(t, is.Equal(len(argv), 3))
		assert.Assert(t, is.Equal(argv[0], 1))
		assert.Assert(t, is.Equal(argv[1], 2))
		assert.Assert(t, is.Equal(argv[2], 3))

		isGetTest1 = true
	})
	event.on("test", func(argv ...interface{}) {
		// fmt.Println("test2", argv)
		assert.Assert(t, is.Equal(len(argv), 3))
		assert.Assert(t, is.Equal(argv[0], 1))
		assert.Assert(t, is.Equal(argv[1], 2))
		assert.Assert(t, is.Equal(argv[2], 3))
		isGetTest2 = true
	})

	eventSub.on("test", func(argv ...interface{}) {
		isGetTest3 = true
	})

	event.emit("test", 1, 2, 3)
	event.emit("test1", 1, 2, 3)

	assert.Assert(t, is.Equal(isGetTest1, true))
	assert.Assert(t, is.Equal(isGetTest2, true))
	assert.Assert(t, is.Equal(isGetTest3, false))
}

func Test_off(t *testing.T) {
	event := Event{}

	event.off("test")
}
