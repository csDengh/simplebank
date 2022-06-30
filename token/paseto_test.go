package token

import (
	"fmt"
	"testing"
	"time"

	"github.com/csdengh/cur_blank/utils"
	"github.com/stretchr/testify/require"
)

func TestPaseto(t *testing.T) {

	pm, err := NewPasetoMaker(utils.RandomString(32))
	require.NoError(t, err)
	require.NotEmpty(t, pm)

	username := utils.RandomOwner()
	timeDur := time.Minute

	token, pl, err := pm.CreateToken(username, timeDur)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, pl)

	plNew, err := pm.ValidToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, plNew)
	fmt.Println(pl)
	fmt.Println(plNew)
	require.Equal(t, pl.Username, plNew.Username)
	require.WithinDuration(t, pl.TimeExprieAt, plNew.TimeExprieAt, time.Millisecond)
	require.WithinDuration(t, pl.IssueAt, plNew.IssueAt, time.Millisecond)
}

func TestPasetoExpire(t *testing.T) {
	pm, err := NewPasetoMaker(utils.RandomString(32))
	require.NoError(t, err)
	require.NotEmpty(t, pm)

	username := utils.RandomOwner()
	timeDur := time.Millisecond

	token, pl, err := pm.CreateToken(username, timeDur)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, pl)

	time.Sleep(time.Millisecond)
	_, err = pm.ValidToken(token)
	require.Error(t, err)
}
