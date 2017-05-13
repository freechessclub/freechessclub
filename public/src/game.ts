// Copyright 2017 The Free Chess Club.

// An online chess game
export const game = {
  captured: {},
  chess: null,
  color: '',
  history: {moves: [], chess: null, id: -1},
  premove: null,
  bclock: null,
  btime: 0,
  wclock: null,
  wtime: 0,
};

export default game;
