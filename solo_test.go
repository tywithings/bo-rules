package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSoloRulesetInterface(t *testing.T) {
	var _ Ruleset = (*SoloRuleset)(nil)
}

func TestSoloName(t *testing.T) {
	r := SoloRuleset{}
	require.Equal(t, "solo", r.Name())
}

func TestSoloCreateNextBoardStateSanity(t *testing.T) {
	boardState := &BoardState{}
	r := SoloRuleset{}
	_, err := r.CreateNextBoardState(boardState, []SnakeMove{})
	require.NoError(t, err)
}

func TestSoloIsGameOver(t *testing.T) {
	tests := []struct {
		Snakes   []Snake
		Expected bool
	}{
		{[]Snake{}, true},
		{[]Snake{{}}, false},
		{[]Snake{{}, {}, {}}, false},
		{[]Snake{{EliminatedCause: EliminatedByOutOfBounds}}, true},
		{
			[]Snake{
				{EliminatedCause: EliminatedByOutOfBounds},
				{EliminatedCause: EliminatedByOutOfBounds},
				{EliminatedCause: EliminatedByOutOfBounds},
			},
			true,
		},
	}

	r := SoloRuleset{}
	for _, test := range tests {
		b := &BoardState{
			Height: 11,
			Width:  11,
			Snakes: test.Snakes,
			Food:   []Point{},
		}

		actual, err := r.IsGameOver(b)
		require.NoError(t, err)
		require.Equal(t, test.Expected, actual)
	}
}

// Checks that a single snake doesn't end the game
// also that:
// - snake moves okay
// - food gets consumed
// - snake grows and gets health from food
var soloCaseNotOver = gameTestCase{
	"Solo Case Game Not Over",
	&BoardState{
		Width:  10,
		Height: 10,
		Snakes: []Snake{
			{
				ID:     "one",
				Body:   []Point{{1, 1}, {1, 2}},
				Health: 100,
			},
		},
		Food:    []Point{{0, 0}, {1, 0}},
		Hazards: []Point{},
	},
	[]SnakeMove{
		{ID: "one", Move: MoveDown},
	},
	nil,
	&BoardState{
		Width:  10,
		Height: 10,
		Snakes: []Snake{
			{
				ID:     "one",
				Body:   []Point{{1, 0}, {1, 1}, {1, 1}},
				Health: 100,
			},
		},
		Food:    []Point{{0, 0}},
		Hazards: []Point{},
	},
}

func TestSoloCreateNextBoardState(t *testing.T) {
	cases := []gameTestCase{
		// inherits these test cases from standard
		standardCaseErrNoMoveFound,
		standardCaseErrZeroLengthSnake,
		standardCaseMoveEatAndGrow,
		standardMoveAndCollideMAD,
		soloCaseNotOver,
	}
	r := SoloRuleset{}
	rb := NewRulesetBuilder().WithParams(map[string]string{
		ParamGameType: GameTypeSolo,
	})
	for _, gc := range cases {
		gc.requireValidNextState(t, &r)
		// also test a RulesBuilder constructed instance
		gc.requireValidNextState(t, rb.Ruleset())
		// also test a pipeline with the same settings
		gc.requireValidNextState(t, NewRulesetBuilder().PipelineRuleset(GameTypeSolo, NewPipeline(soloRulesetStages...)))
	}
}

// Test a snake running right into the wall is properly eliminated
func TestSoloEliminationOutOfBounds(t *testing.T) {
	r := SoloRuleset{}

	// Using MaxRand is important because it ensures that the snakes are consistently placed in a way this test will work.
	// Actually random placement could result in the assumptions made by this test being incorrect.
	initialState, err := CreateDefaultBoardState(MaxRand, 2, 2, []string{"one"})
	require.NoError(t, err)

	_, next, err := r.Execute(
		initialState,
		r.Settings(),
		[]SnakeMove{{ID: "one", Move: "right"}},
	)
	require.NoError(t, err)
	require.NotNil(t, initialState)

	ended, next, err := r.Execute(
		next,
		r.Settings(),
		[]SnakeMove{{ID: "one", Move: "right"}},
	)
	require.NoError(t, err)
	require.NotNil(t, initialState)

	require.True(t, ended)
	require.Equal(t, EliminatedByOutOfBounds, next.Snakes[0].EliminatedCause)
	require.Equal(t, "", next.Snakes[0].EliminatedBy)
	require.Equal(t, 1, next.Snakes[0].EliminatedOnTurn)
}
