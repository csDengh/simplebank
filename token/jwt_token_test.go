package token

import (
	"fmt"
	"testing"
	"time"

	"github.com/csdengh/cur_blank/utils"
	"github.com/stretchr/testify/require"
)

func TestJwtToken(t *testing.T) {
	jm, err := NewJwtMaker(utils.RandomString(32))
	require.NoError(t, err)
	require.NotEmpty(t, jm)

	username := utils.RandomOwner()
	timeDur := time.Minute

	token, pl, err := jm.CreateToken(username, timeDur)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, pl)

	plNew, err := jm.ValidToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, plNew)
	fmt.Println(pl)
	fmt.Println(plNew)
	require.Equal(t, pl.Username, plNew.Username)
	require.WithinDuration(t, pl.TimeExprieAt, plNew.TimeExprieAt, time.Millisecond)
	require.WithinDuration(t, pl.IssueAt, plNew.IssueAt, time.Millisecond)

}

func TestJwtExpire(t *testing.T) {

	jm, err := NewJwtMaker(utils.RandomString(32))
	require.NoError(t, err)
	require.NotEmpty(t, jm)

	username := utils.RandomOwner()
	timeDur := time.Millisecond

	token, pl, err := jm.CreateToken(username, timeDur)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, pl)

	time.Sleep(time.Millisecond)
	_, err = jm.ValidToken(token)
	require.Error(t, err)
}

func TestJwtInvalid(t *testing.T){
	jm, err := NewJwtMaker(utils.RandomString(32))
	require.NoError(t, err)
	require.NotEmpty(t, jm)

	username := utils.RandomOwner()
	timeDur := time.Minute

	token, pl, err := jm.CreateToken(username, timeDur)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, pl)

	token = utils.RandomString(50)
	
	_, err = jm.ValidToken(token)
	require.Error(t, err)
}
