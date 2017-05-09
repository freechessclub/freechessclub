// Copyright 2017 Abhishek Kulkarni

/**
 * Convert seconds to HH:MM:SS.
 */
export function SToHHMMSS(sec: number) {
  var h = Math.abs(Math.floor(sec / 3600));
  var m = Math.abs(Math.floor(sec % 3600 / 60));
  var s = Math.abs(Math.floor(sec % 3600 % 60));
  return ((sec < 0 ? '-' : '')
  + (h > 0 ? (h >= 0 && h < 10 ? '0' : '') + h + ':' : '')
  + (m >= 0 && m < 10 ? '0' : '') + m + ':'
  + (s >= 0 && s < 10 ? '0' : '') + s);
}

export function startBlackClock(game, clock) {
  return setInterval(function() {
    if (game.chess.turn() === 'w') {
      return;
    }

    game.btime = game.btime - 1;
    if (game.btime < 20 && clock.css('color') !== 'red') {
      clock.css('color', 'red');
    }

    if (game.btime > 20) {
      clock.css('color', '');
    }

    clock.text(SToHHMMSS(game.btime));
  }, 1000);
}

export function startWhiteClock(game, clock) {
  return setInterval(function() {
    if (game.chess.turn() === 'b') {
      return;
    }

    game.wtime = game.wtime - 1;
    if (game.wtime < 20 && clock.css('color') !== 'red') {
      clock.css('color', 'red');
    }

    if (game.wtime > 20) {
      clock.css('color', '');
    }

    clock.text(SToHHMMSS(game.wtime));
  }, 1000);
}
