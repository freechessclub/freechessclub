// Copyright 2017 The Free Chess Club.

import * as $ from 'jquery';

import ReconnectingWebSocket = require('reconnecting-websocket');
import anchorme from 'anchorme';
import 'bootstrap';
import * as Chess from 'chess.js';
import * as ChessBoard from 'chessboardjs';
import * as clock from './clock';
import * as highlight from './highlight';

// A FICS session
const session = {
  connected: false,
  handle: '',
  ws: null,
};

// List of active tabs
let tabs;

// An online chess game
const game = {
  chess: null,
  color: '',
  history: null,
  premove: null,
  bclock: null,
  btime: 0,
  wclock: null,
  wtime: 0,
};

// Type of messages we receive from the proxy
enum MessageType {
  Control = 0,
  ChannelTell,
  PrivateTell,
  GameMove,
  GameStart,
  GameEnd,
  Unknown,
}

/**
 * Show captured piece.
 * @param color
 * @param captured
 */
function showCapture(color, captured) {
  if (typeof captured !== 'undefined') {
    if (color === game.color) {
      $('#opponent-captured').append(captured);
    } else {
      $('#player-captured').append(captured);
    }
  }
}

const onDragStart = (source, piece, position, orientation) => {
  const chess = game.chess;
  if (chess === null) {
    return false;
  }

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

function movePlayer(source, target) {
  const chess = game.chess;
  // see if the move is legal
  const move = chess.move({
    from: source,
    to: target,
    promotion: 'q', // TODO: Allow non-queen promotes
  });

  // illegal move
  if (move === null) {
    highlight.unHighlightSquare();
    return 'snapback';
  }

  session.ws.send(JSON.stringify({ type: MessageType.Control, command: 0, text: source + '-' + target }));
  highlight.highlightMove(move.from, move.to);
  showCapture(move.color, move.captured);
  highlight.showCheck(move.color, move.san);
}

const onDrop = (source, target) => {
  // premove if it is not my turn yet
  if (game.color !== game.chess.turn()) {
    game.premove = {source, target};
    return highlight.highlightPreMove(source, target);
  } else {
    return movePlayer(source, target);
  }
};

// update the board position after the piece snap
// for castling, en passant, pawn promotion
const onSnapEnd = () => {
  board.position(game.chess.fen());
};

// Chess board
const board: any = ChessBoard('board', {
  position: 'start',
  showNotation: true,
  draggable: true,
  onDragStart,
  onDrop,
  onSnapEnd,
  pieceTheme: 'assets/img/chesspieces/wikipedia-svg/{piece}.svg',
});

// enable tooltips
$(() => {
  $('[data-toggle="tooltip"]').tooltip();
});

// Allow chat card to be collapsed
$('#collapse-chat').on('hidden.bs.collapse', () => {
  $('#chat-toggle-icon').removeClass('fa-toggle-up').addClass('fa-toggle-down');
});
$('#collapse-chat').on('show.bs.collapse', () => {
  $('#chat-toggle-icon').removeClass('fa-toggle-down').addClass('fa-toggle-up');
});

jQuery(document.body).on('click', '.closeTab', (event) => {
  const tabContentId = $(this).parent().attr('href');
  $(this).parent().remove();
  delete tabs['content-' + tabContentId.substr(1)];
  $('#tabs a:last').tab('show');
  $(tabContentId).remove();
});

$(document).on('shown.bs.tab', 'a[data-toggle="tab"]', (e) => {
  const tab = $(e.target);
  tab.css('color', 'black');
});

function handleChatMsg(from, data) {
  let tab;
  if (!tabs.hasOwnProperty(from)) {
    let chName = from;
    if (from === '4') {
      chName = 'Help';
    }
    $('<a class="flex-sm-fill text-sm-center nav-link" data-toggle="tab" href="#content-' +
      from + '" id="' + from + '" role="tab">' + chName +
      '<span class="btn btn-default btn-sm closeTab">×</span></a>').appendTo('#tabs');
    $('<div class="tab-pane chat-text" id="content-' + from + '" role="tabpanel"></div>').appendTo('.tab-content');
    $('.chat-text').height($('#board').height() - 40);
    tab = $('#content-' + from);
    tabs[from] = tab;
  } else {
    tab = tabs[from];
  }

  let who = '';
  let tabheader = $('#' + $('ul#tabs a.active').attr('id'));
  if (data.hasOwnProperty('handle')) {
    let textclass = '';
    if (session.handle === data.handle) {
      textclass = ' class="mine"';
    }
    who = '<strong' + textclass + '>' + $('<span/>').text(data.handle).html() + '</strong>: ';
    if (data.type === MessageType.ChannelTell) {
      tabheader = $('#' + data.channel);
    } else {
      tabheader = $('#' + data.handle);
    }
  }
  tab.append(who +
    anchorme($('<span/>').text(data.text).html(), {attributes: [{name: 'target', value: '_blank'} ]}) + '</br>');

  if (tabheader.hasClass('active')) {
    tab.scrollTop(tab[0].scrollHeight);
  } else {
    tabheader.css('color', 'red');
  }
}

function handleICSMsg(message) {
  const data = JSON.parse(message.data);
  switch (data.type) {
    case MessageType.Control:
      if (session.connected === false && data.command === 1) {
        session.connected = true;
        session.handle = data.text;
        $('#chat-status').text('Connected as ' + session.handle);
      }
      break;
    case MessageType.ChannelTell:
      handleChatMsg(data.channel, data);
      break;
    case MessageType.PrivateTell:
      handleChatMsg(data.handle, data);
      break;
    case MessageType.GameMove:
      game.btime = data.btime;
      game.wtime = data.wtime;

      // role 1: I am playing and it is my move
      // role -1: I am playing and it is my opponent's move
      if (game.chess === null) {
        game.chess = Chess();
        board.start(false);
        game.history = {moves: [], chess: null, id: -1};
        $('#player-captured').text('');
        $('#opponent-captured').text('');
        if (data.role === 1) {
          game.color = 'w';
          board.orientation('white');
          game.wclock = clock.startWhiteClock(game, $('#player-time'));
          game.bclock = clock.startBlackClock(game, $('#opponent-time'));
          $('#player-name').text(data.wname);
          $('#opponent-name').text(data.bname);
        } else if (data.role === -1) {
          game.color = 'b';
          board.orientation('black');
          game.bclock = clock.startBlackClock(game, $('#player-time'));
          game.wclock = clock.startWhiteClock(game, $('#opponent-time'));
          $('#player-name').text(data.bname);
          $('#opponent-name').text(data.wname);
        }
      }

      if (data.role === 1) {
        if (data.move !== 'none') {
          const move = game.chess.move(data.move);
          if (move !== null) {
            highlight.highlightMove(move.from, move.to);
            showCapture(move.color, move.captured);
            highlight.showCheck(move.color, move.san);
          }
          if (game.premove !== null) {
            movePlayer(game.premove.source, game.premove.target);
            game.premove = null;
          }
        }
      }
      board.position(data.fen);
      break;
    case MessageType.GameStart:
      break;
    case MessageType.GameEnd:
      clearInterval(game.wclock);
      clearInterval(game.bclock);
      displayHistory();
      delete game.chess;
      game.chess = null;
      break;
    case MessageType.Unknown:
    default:
      handleChatMsg($('ul#tabs a.active').attr('id'), data);
      break;
  }
}

function disconnectICS() {
  $('#chat-status').text('Disconnected');
  session.connected = false;
  session.handle = '';
}

function connectToICS(user?: string, pass?: string) {
  const login = (typeof user !== 'undefined' && typeof pass !== 'undefined');
  let loginOptions = '';
  if (login) {
    loginOptions += '?login=1';
  }
  const conn = new ReconnectingWebSocket(
    location.protocol.replace('http', 'ws') + '//' + location.host + '/ws' + loginOptions);
  conn.onmessage = handleICSMsg;
  conn.onclose = disconnectICS;
  if (login) {
    conn.onopen = () => {
      conn.send(JSON.stringify({ type: MessageType.Control, command: 1, text: '[' + user + ',' + btoa(pass) + ']' }));
    };
  }
  return conn;
}

$('#input-form').on('submit', (event) => {
  event.preventDefault();
  let text;
  if (!$('#input-command').is(':checked')) {
    if ($('#input-text').val().charAt(0) !== '@') {
      const msg = $('#input-text').val();
      const tab = $('ul#tabs a.active').attr('id');
      text = 't ' + tab + ' ' + msg;
      handleChatMsg(tab, { type: MessageType.ChannelTell, channel: tab, handle: session.handle, text: msg });
    } else {
      text = $('#input-text').val().substr(1);
    }
  } else {
    if ($('#input-text').val().charAt(0) !== '@') {
      text = $('#input-text').val();
    } else {
      text = $('#input-text').val().substr(1);
    }
  }
  session.ws.send(JSON.stringify({ type: MessageType.Control, command: 0, text }));
  $('#input-text').val('');
});

$(document).ready(() => {
  session.ws = connectToICS();
  session.connected = false;
  $('#chat-status').text('Connecting...');
  $('#opponent-time').text('00:00');
  $('#player-time').text('00:00');
  $('.chat-text').height($('#board').height() - 40);
  tabs = { 53: $('#content-53') };
  board.start(false);
  game.history = {moves: [], chess: null, id: -1};
});

function displayHistory() {
  if (game.history.chess === null) {
    game.history.chess = Chess();
  }

  // refresh history
  if (game.chess !== null) {
    const moves = game.chess.history();
    if (game.history.moves.length < moves.length) {
      for (let i = game.history.moves.length - 1; i < moves.length; i++) {
        game.history.chess.move(moves[i]);
        game.history.moves.push(game.history.chess.fen());
      }
    }
  }

  if (game.history.id < 0) {
    game.history.id = game.history.moves.length - 1;
  }

  board.position(game.history.moves[game.history.id]);
}

$('#fast-backward').on('click', (event) => {
  game.history.id = 0;
  displayHistory();
});

$('#backward').on('click', (event) => {
  if (game.history.id > 0) {
    game.history.id = game.history.id - 1;
  }
  displayHistory();
});

$('#forward').on('click', (event) => {
  if (game.history.id < game.history.moves.length - 1) {
    game.history.id = game.history.id + 1;
  }
  displayHistory();
});

$('#fast-forward').on('click', (event) => {
  game.history.id = game.history.moves.length - 1;
  displayHistory();
});

$('#resign').on('click', (event) => {
  if (game.chess !== null && session.ws !== null) {
    session.ws.send(JSON.stringify({ type: MessageType.Control, command: 0, text: 'resign' }));
  }
});

$('#abort').on('click', (event) => {
  if (game.chess !== null && session.ws !== null) {
    session.ws.send(JSON.stringify({ type: MessageType.Control, command: 0, text: 'abort' }));
  }
});

$('#takeback').on('click', (event) => {
  if (game.chess !== null && session.ws !== null) {
    if (game.chess.turn() === game.color) {
      session.ws.send(JSON.stringify({ type: MessageType.Control, command: 0, text: 'take 2'}));
    } else {
      session.ws.send(JSON.stringify({ type: MessageType.Control, command: 0, text: 'take 1'}));
    }
  }
});

$('#draw').on('click', (event) => {
  if (game.chess !== null && session.ws !== null) {
    session.ws.send(JSON.stringify({ type: MessageType.Control, command: 0, text: 'draw' }));
  }
});

$('#disconnect').on('click', (event) => {
  $('#chat-status').text('Disconnecting...');
  session.ws.close();
});

$('#login').on('click', (event) => {
  $('#chat-status').text('Connecting...');
  const user = $('#login-user').val();
  const pass = $('#login-pass').val();
  session.ws = connectToICS(user, pass);
  $('#login-screen').modal('hide');
});

$('#connect-user').on('click', (event) => {
  if (session.connected !== true) {
    $('#login-screen').modal('show');
  }
});

$('#connect-guest').on('click', (event) => {
  if (session.connected !== true) {
    $('#chat-status').text('Connecting...');
    session.ws = connectToICS();
  }
});

$(window).focus(() => {
  if (game.chess !== null) {
    board.position(game.chess.fen(), false);
  }
});

$(window).resize(() => {
  board.resize();
  $('.chat-text').height($('#board').height() - 40);
});
