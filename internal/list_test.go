package internal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	l := NewList[string, string](5, LIST)
	require.Equal(t, uint(5), l.capacity)
	require.Equal(t, LIST, l.listType)
	for i := 0; i < 5; i++ {
		evicted := l.PushFront(NewEntry[string, string](fmt.Sprintf("%d", i), "", 1, 0))
		require.Nil(t, evicted)
	}
	require.Equal(t, 5, l.len)
	require.Equal(t, "4/3/2/1/0", l.display(LIST))
	require.Equal(t, "0/1/2/3/4", l.displayReverse(LIST))

	evicted := l.PushFront(NewEntry[string, string]("5", "", 1, 0))
	require.Equal(t, "0", evicted.key)
	require.Equal(t, 5, l.len)
	require.Equal(t, "5/4/3/2/1", l.display(LIST))
	require.Equal(t, "1/2/3/4/5", l.displayReverse(LIST))

	for i := 0; i < 5; i++ {
		entry := l.PopTail()
		require.Equal(t, fmt.Sprintf("%d", i+1), entry.key)
	}
	entry := l.PopTail()
	require.Nil(t, entry)

}

func TestWheelList(t *testing.T) {
	l := NewList[string, string](5, WHEEL_LIST)
	require.Equal(t, uint(5), l.capacity)
	require.Equal(t, WHEEL_LIST, l.listType)
	for i := 0; i < 5; i++ {
		evicted := l.PushFront(NewEntry[string, string](fmt.Sprintf("%d", i), "", 1, 0))
		require.Nil(t, evicted)
	}
	require.Equal(t, 5, l.len)
	require.Equal(t, "4/3/2/1/0", l.display(WHEEL_LIST))
	require.Equal(t, "0/1/2/3/4", l.displayReverse(WHEEL_LIST))

	evicted := l.PushFront(NewEntry[string, string]("5", "", 1, 0))
	require.Equal(t, "0", evicted.key)
	require.Equal(t, 5, l.len)
	require.Equal(t, "5/4/3/2/1", l.display(WHEEL_LIST))
	require.Equal(t, "1/2/3/4/5", l.displayReverse(WHEEL_LIST))

	for i := 0; i < 5; i++ {
		entry := l.PopTail()
		require.Equal(t, fmt.Sprintf("%d", i+1), entry.key)
	}
	entry := l.PopTail()
	require.Nil(t, entry)

}
