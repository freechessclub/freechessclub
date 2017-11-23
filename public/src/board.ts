// Copyright 2017 Free Chess Club.

import * as ChessBoard from 'oakmac-chessboard';

import game from './game';
import * as highlight from './highlight';
import { movePiece } from './index';

const onDragStart = (source, piece, position, orientation) => {
  if (!game.chess) {
    return false;
  }

  // stop dragging if the game is over or if it's opponents piece or if we are observing
  if (game.chess.game_over() || (game.color !== piece.charAt(0)) || game.obs) {
    return false;
  }

  if (game.premove !== null) {
    highlight.unHighlightSquare(game.premove.source);
    highlight.unHighlightSquare(game.premove.target);
    game.premove = null;
  }

  // get list of possible moves for this square
  const moves = game.chess.moves({square: source, verbose: true});
  highlight.highlightSquare(source);
  for (const move of moves) {
    highlight.highlightSquare(move.to);
  }
};

const onDrop = (source, target) => {
  if (!game.chess) {
    return;
  }
  // premove if it is not my turn yet
  if (game.color !== game.chess.turn() && source !== target) {
    game.premove = {source, target};
    return highlight.highlightPreMove(source, target);
  } else {
    return movePiece(source, target);
  }
};

// update the board position after the piece snap
// for castling, en passant, pawn promotion
const onSnapEnd = () => {
  if (!game.chess) {
    return;
  }
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
  pieceTheme: 'www/img/chesspieces/wikipedia-svg/{piece}.svg',
});

export default board;
