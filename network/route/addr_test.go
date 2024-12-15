package route

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddr(t *testing.T) {
	addr, err := ParseAddr("1.2.3.4/23")
	require.NoError(t, err)
	assert.Equal(t, "1.2.2.0", addr.IP.String())
	assert.Equal(t, "255.255.254.0", net.IP(addr.Mask).String())

	addr, err = ParseAddr("1.2.3.4")
	require.NoError(t, err)
	assert.Equal(t, "1.2.3.4", addr.IP.String())
	assert.Nil(t, addr.Mask)
}
