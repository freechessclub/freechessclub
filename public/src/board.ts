// Copyright 2017 The Free Chess Club.

import * as ChessBoard from 'chessboardjs';

import game from './game';
import * as highlight from './highlight';
import { movePiece } from './index';

const onDragStart = (source, piece, position, orientation) => {
  const chess = game.chess;
  if (chess === null) {
    return false;
  }

  // stop dragging if the game is over or if it's opponents piece
  if (chess.game_over() || (game.color !== piece.charAt(0))) {
    return false;
  }

  if (game.premove !== null) {
    highlight.unHighlightSquare(game.premove.source);
    highlight.unHighlightSquare(game.premove.target);
    game.premove = null;
  }

  // get list of possible moves for this square
  const moves = chess.moves({square: source, verbose: true});
  highlight.highlightSquare(source);
  for (const move of moves) {
    highlight.highlightSquare(move.to);
  }
};

const onDrop = (source, target) => {
  // premove if it is not my turn yet
  if (game.color !== game.chess.turn()) {
    game.premove = {source, target};
    return highlight.highlightPreMove(source, target);
  } else {
    return movePiece(source, target);
  }
};

// update the board position after the piece snap
// for castling, en passant, pawn promotion
const onSnapEnd = () => {
  board.position(game.chess.fen());
};

// Chess board
export const board: any = ChessBoard('board', {
  position: 'start',
  showNotation: true,
  draggable: true,
  onDragStart,
  onDrop,
  onSnapEnd,
  pieceTheme: 'assets/img/chesspieces/wikipedia-svg/{piece}.svg',
});

export default board;
