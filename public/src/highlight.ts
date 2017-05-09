// Copyright 2017 The Free Chess Club.

/**
 * Highlight an arbitrary square (to show possible moves or the last move)
 * @param square Square to highlight
 */
export function highlightSquare(square?: string): void {
  if (square === undefined) {
    return;
  }
  const e = $('#board .square-' + square);
  if (e.hasClass('black-3c85d')) {
    e.css('background', '#278881');
  } else {
    e.css('background', '#e6ffdd');
  }
}

/**
 * Remove highlights from a square or all squares on the board.
 * @param square Square to unhighlight
 */
export function unHighlightSquare(square?: string): void {
  if (square) {
    $('#board .square-' + square).css('background', '');
  } else {
    $('#board .square-55d63').css('background', '');
  }
}

/**
 * Highlight a square to show an active check.
 * @param square Square to highlight
 */
export function highlightCheck(square?: string): void {
  if (square === undefined) {
    return;
  }
  const e = $('#board .square-' + square);
  if (e.hasClass('black-3c85d')) {
    e.css('background', '#aa8881');
  } else {
    e.css('background', '#ffdddd');
  }
}

/**
 * Highlight a move. This is used to provide a visual cue to display the
 * previous move on the board.
 * @param source The source square to highlight
 * @param target The target square to highlight
 */
export function highlightMove(source: string, target: string): void {
  unHighlightSquare();
  highlightSquare(source);
  highlightSquare(target);
}

/**
 * Highlight a premove. Premoves are highlighted with a different color than the last played move.
 * @param source The source square to highlight
 * @param target The target square to highlight
 */
export function highlightPreMove(source: string, target: string): void {
  highlightCheck(source);
  highlightCheck(target);
}

/**
 * Function to swap the color: B -> W or W -> B.
 */
export function swapColor(color: string): string {
  return (color === 'w') ? 'b' : 'w';
}

/**
 * Show check on the board by highlighting the king square.
 */
export function showCheck(color: string, san: string) {
  if (san.slice(-1) === '+') {
    const square = $('div').find("[data-piece='" + swapColor(color) + "K']");
    highlightCheck(square.parent().data('square'));
  }
}
