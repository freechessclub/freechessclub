// Copyright 2017 The Free Chess Club.

import anchorme from 'anchorme';
import 'bootstrap';
import * as Chess from 'chess.js';
import * as $ from 'jquery';

import board from './board';
import channelList from './channels';
import * as clock from './clock';
import game from './game';
import * as highlight from './highlight';
import History from './history';
import MessageType from './message';
import Session from './session';
import * as Sounds from './sounds';

// ICS session
let session: Session;

// toggle game sounds
let soundToggle: boolean = true;

// pending takeback requests
let pendingTakeback = 0;

// list of active tabs
const tabsList = {
  console: $('#content-console'),
};

function showCapturePiece(color: string, piece: string): void {
  const p: string = highlight.swapColor(color) + piece.toUpperCase();
  if (game.color === color) {
    if (game.playerCaptured[p] === undefined) {
      game.playerCaptured[p] = 0;
    }
    game.playerCaptured[p]++;
  } else {
    if (game.oppCaptured[p] === undefined) {
      game.oppCaptured[p] = 0;
    }
    game.oppCaptured[p]++;
  }

  $('#player-captured').empty();
  $('#opponent-captured').empty();
  for (const key in game.playerCaptured) {
    if (game.playerCaptured.hasOwnProperty(key)) {
      $('#player-captured').append(
        '<img id="' + key + '" src="assets/img/chesspieces/wikipedia-svg/' +
          key + '.svg"/><small>' + game.playerCaptured[key] + '</small>');
    }
  }
  for (const key in game.oppCaptured) {
    if (game.oppCaptured.hasOwnProperty(key)) {
      $('#opponent-captured').append(
        '<img id="' + key + '" src="assets/img/chesspieces/wikipedia-svg/' +
          key + '.svg"/><small>' + game.oppCaptured[key] + '</small>');
    }
  }
}

(window as any).showMove = (id: number) => {
  if (game.history) {
    game.history.display(id);
  }
};

function addMoveHistory(move: any): void {
  const id: number = game.history.length();
  if (id % 2 === 1) {
    $('#move-history').append('<tr><td><div class="moveNumber">'
      + (id + 1) / 2 + '.</div><a href="javascript:void(0);" onclick="showMove(' + id + ')">'
      + move.san + '</a></td><td>&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;</td></tr>');
    $('#left-panel').scrollTop(document.getElementById('left-panel').scrollHeight);
  } else {
    $('#move-history tr:last td').eq(1).html('<a href="javascript:void(0);" onclick="showMove(' +
      id + ')">' + move.san + '</a>');
  }
}

export function movePiece(source, target) {
  // see if the move is legal
  const move = game.chess.move({
    from: source,
    to: target,
    promotion: 'q', // TODO: Allow non-queen promotes
  });

  // illegal move
  if (move === null) {
    highlight.unHighlightSquare();
    return 'snapback';
  }

  session.send({ type: MessageType.Control, command: 0, text: source + '-' + target });
  game.history.add(game.chess.fen());
  addMoveHistory(move);
  highlight.highlightMove(move.from, move.to);
  if (move.captured) {
    showCapturePiece(move.color, move.captured);
  }
  if (highlight.showCheck(move.color, move.san)) {
    if (soundToggle) {
      Sounds.checkSound.play();
    }
  } else {
    if (soundToggle) {
      if (move.captured) {
        Sounds.captureSound.play();
      } else {
        Sounds.moveSound.play();
      }
    }
  }
}

// enable tooltips
$(() => {
  $('[data-toggle="tooltip"]').tooltip();
});

// Allow chat card to be collapsed
$('#collapse-chat').on('hidden.bs.collapse', () => {
  $('#chat-toggle-icon').removeClass('fa-toggle-up').addClass('fa-toggle-down');
});
$('#collapse-chat').on('shown.bs.collapse', () => {
  $('#chat-toggle-icon').removeClass('fa-toggle-down').addClass('fa-toggle-up');
});

$('#collapse-history').on('hidden.bs.collapse', () => {
  $('#history-toggle-icon').removeClass('fa-toggle-up').addClass('fa-toggle-down');
});
$('#collapse-history').on('shown.bs.collapse', () => {
  $('#history-toggle-icon').removeClass('fa-toggle-down').addClass('fa-toggle-up');
});

$(document.body).on('click', '.closeTab', (event) => {
  const tabContentId: string = $(event.target).parent().attr('id');
  $(event.target).parent().remove();
  delete tabsList[tabContentId];
  $('#tabs a:last').tab('show');
  $('#content-' + tabContentId).remove();
});

$(document).on('shown.bs.tab', 'a[data-toggle="tab"]', (e) => {
  const tab = $(e.target);
  tab.css('color', 'black');
});

function showStatusMsg(msg: string) {
  $('#game-status').html(msg + '<br/>');
  $('#left-panel').scrollTop(document.getElementById('left-panel').scrollHeight);
}

function createOrGetTab(from) {
  if (tabsList.hasOwnProperty(from)) {
    return tabsList[from];
  }

  let chName = from;
  if (channelList[from] !== undefined) {
    chName = channelList[from];
  }

  $('<a class="flex-sm-fill text-sm-center nav-link" data-toggle="tab" href="#content-' +
    from + '" id="' + from + '" role="tab">' + chName +
    '<span class="btn btn-default btn-sm closeTab">×</span></a>').appendTo('#tabs');
  $('<div class="tab-pane chat-text" id="content-' + from + '" role="tabpanel"></div>').appendTo('.tab-content');
  $('.chat-text').height($('#board').height() - 40);
  tabsList[from] = $('#content-' + from);
  return tabsList[from];
}

function handleChatMsg(from, data) {
  const tab = createOrGetTab(from);
  let who = '';
  if (data.hasOwnProperty('handle')) {
    let textclass = '';
    if (session.getHandle() === data.handle) {
      textclass = ' class="mine"';
    }
    who = '<strong' + textclass + '>' + $('<span/>').text(data.handle).html() + '</strong>: ';
  }

  let tabheader;
  if (data.type === MessageType.ChannelTell) {
    tabheader = $('#' + data.channel);
  } else if (data.type === MessageType.PrivateTell) {
    tabheader = $('#' + data.handle);
  } else {
    tabheader = $('#console');

    let takeBacker = null;
    let action = null;
    if (pendingTakeback) {
      let takebackMatches = data.text.match(/(\w+) (\w+) the takeback request\./);
      if (takebackMatches !== null && takebackMatches.length > 1) {
        takeBacker = takebackMatches[1];
        action = takebackMatches[2];
      } else {
        takebackMatches = data.text.match(/You (\w+) the takeback request from (\w+)\./);
        if (takebackMatches !== null && takebackMatches.length > 1) {
          takeBacker = takebackMatches[2];
          action = takebackMatches[1];
        }
      }

      if (takeBacker !== null && action !== null) {
        if (takeBacker === $('#opponent-name').text()) {
          if (action.startsWith('decline')) {
            pendingTakeback = 0;
            return;
          }
          game.premove = null;
          for (let i = 0; i < pendingTakeback; i++) {
            if (game.chess) {
              game.chess.undo();
            }
            if (game.history) {
              game.history.undo();
            }
          }
          pendingTakeback = 0;
          return;
        }
      }
    }

    const takebackReq = data.text.match(/(\w+) would like to take back (\d+) half move\(s\)\./);
    if (takebackReq != null && takebackReq.length > 1) {
      if (takebackReq[1] === $('#opponent-name').text()) {
        pendingTakeback = Number(takebackReq[2]);
      }
    }

    const chListMatches = data.text.match(/-- channel list: \d+ channels --(?:\n)([\d\s]*)/);
    if (chListMatches !== null && chListMatches.length > 1) {
      $('#chan-dropdown-menu').empty();
      const joinedChannels = chListMatches[1].split(/\s+/);
      joinedChannels.forEach((ch) => {
        $('#chan-dropdown-menu').append(
          '<a class="dropdown-item noselect" id="ch-' + ch +
          '">' + channelList[Number(ch)] + '</a>');

        $('#ch-' + ch).on('click', (event) => {
          event.preventDefault();
          createOrGetTab(Number(ch));
          // createOrGetTab(Number(ch)).tab('show');
        });
      });
      return;
    }
    if (
      data.text === 'Style 12 set.' ||
      data.text === 'You will not see seek ads.' ||
      data.text === 'You will now hear communications echoed.'
    ) {
      return;
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

function ICSMessageHandler(message) {
  const data = JSON.parse(message.data);
  if (Array.isArray(data)) {
    data.map((m) => messageHandler(m));
  } else {
    messageHandler(data);
  }
}

function messageHandler(data) {
  if (data === undefined || data === null) {
    return;
  }

  switch (data.type) {
    case MessageType.Control:
      if (!session.isConnected() && data.command === 1) {
        session.setHandle(data.text);
        session.send({ type: MessageType.Control, command: 0, text: '=ch' });
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

      if (game.chess === null) {
        game.chess = Chess();
        board.start(false);
        game.history = new History(board, game.chess.fen());
        game.playerCaptured = {};
        game.oppCaptured = {};
        $('#player-captured').text('');
        $('#opponent-captured').text('');
        showStatusMsg('');
        $('#move-history').empty();
        $('#player-status').css('background-color', '');
        $('#opponent-status').css('background-color', '');

        // role 0: I am observing
        // role 1: I am playing and it is NOW my move
        if (data.role >= 0) {
          game.color = 'w';
          board.orientation('white');
          game.wclock = clock.startWhiteClock(game, $('#player-time'));
          game.bclock = clock.startBlackClock(game, $('#opponent-time'));
          $('#player-name').text(data.wname);
          $('#opponent-name').text(data.bname);
          if (data.role === 0) {
            game.obs = true;
          }
        // role -1: I am playing and it is NOW my opponent's move
        } else if (data.role === -1) {
          game.color = 'b';
          board.orientation('black');
          game.bclock = clock.startBlackClock(game, $('#player-time'));
          game.wclock = clock.startWhiteClock(game, $('#opponent-time'));
          $('#player-name').text(data.bname);
          $('#opponent-name').text(data.wname);
        }
      }

      if (data.role >= 0) {
        if (data.move !== 'none') {
          const move = game.chess.move(data.move);
          if (move !== null) {
            highlight.highlightMove(move.from, move.to);
            if (move.captured) {
              showCapturePiece(move.color, move.captured);
            }
            if (highlight.showCheck(move.color, move.san)) {
              if (soundToggle) {
                Sounds.checkSound.play();
              }
            } else {
              if (soundToggle) {
                if (move.captured) {
                  Sounds.captureSound.play();
                } else {
                  Sounds.moveSound.play();
                }
              }
            }
            game.history.add(game.chess.fen());
            addMoveHistory(move);
          }

          if (game.premove !== null) {
            movePiece(game.premove.source, game.premove.target);
            game.premove = null;
          }
        }
      }
      board.position(data.fen);
      break;
    case MessageType.GameStart:
      const handle = session.getHandle();
      if (data.playerone === handle) {
        createOrGetTab(data.playertwo);
      } else {
        createOrGetTab(data.playerone);
      }
      break;
    case MessageType.GameEnd:
      if (data.reason < 4 && session.getHandle() === data.winner) {
        // player won
        $('#player-status').css('background-color', '#d4f9d9');
        $('#opponent-status').css('background-color', '#f9d4d4');
      } else if (data.reason < 4 && session.getHandle() === data.loser) {
        // opponent won
        $('#player-status').css('background-color', '#f9d4d4');
        $('#opponent-status').css('background-color', '#d4f9d9');
      } else {
        // tie
        $('#player-status').css('background-color', '#fcddae');
        $('#opponent-status').css('background-color', '#fcddae');
      }

      showStatusMsg(data.message);
      clearInterval(game.wclock);
      clearInterval(game.bclock);
      delete game.chess;
      game.chess = null;
      break;
    case MessageType.Unknown:
    default:
      handleChatMsg('console', data);
      break;
  }
}

function getValue(elt: string): string {
  return $(elt).val() as string;
}

const chatHistory = new Array(15);
let chatHistoryCounter = 0;
let currentCounter = -1;

$('#input-text').keydown((e) => {
  const code = (e.keyCode ? e.keyCode : e.which);
  if (code === 38 && chatHistoryCounter > 0) {
    currentCounter = (currentCounter + 1) % chatHistoryCounter;
    $('#input-text').val(chatHistory[chatHistoryCounter - currentCounter - 1]);
  } else if (code === 40) {
    if (currentCounter > 0 && chatHistoryCounter > 0) {
      currentCounter = (currentCounter - 1) % (chatHistoryCounter - 1);
    }
    $('#input-text').val(chatHistory[chatHistoryCounter - currentCounter - 1]);
  }
});

$('#input-form').on('submit', (event) => {
  event.preventDefault();
  let text;
  const val: string = getValue('#input-text');
  if (val === '' || val === '\n') {
    return;
  }
  chatHistory[chatHistoryCounter] = val;
  chatHistoryCounter = (chatHistoryCounter + 1) % chatHistory.length;
  currentCounter = -1;
  const tab = $('ul#tabs a.active').attr('id');
  if (tab !== 'console') {
    if (val.charAt(0) !== '@') {
      text = 't ' + tab + ' ' + val;
    } else {
      text = val.substr(1);
    }
  } else {
    if (val.charAt(0) !== '@') {
      text = val;
    } else {
      text = val.substr(1);
    }
  }

  session.send({ type: MessageType.Control, command: 0, text });
  $('#input-text').val('');
});

$(document).ready(() => {
  session = new Session(ICSMessageHandler);
  $('#opponent-time').text('00:00');
  $('#player-time').text('00:00');
  $('.chat-text').height($('#board').height() - 40);
  if ($(window).width() < 767) {
    $('#collapse-chat').collapse('hide');
    $('#collapse-history').collapse('hide');
  }
  $('#left-panel').height($('#board').height() - 30);
  board.start(false);
});

$('#fast-backward').on('click', (event) => {
  if (game.history) {
    game.history.beginning();
  }
});

$('#backward').on('click', (event) => {
  if (game.history) {
    game.history.backward();
  }
});

$('#forward').on('click', (event) => {
  if (game.history) {
    game.history.forward();
  }
});

$('#fast-forward').on('click', (event) => {
  if (game.history) {
    game.history.end();
  }
});

$('#resign').on('click', (event) => {
  if (game.chess !== null) {
    session.send({ type: MessageType.Control, command: 0, text: 'resign' });
  } else {
    showStatusMsg('You are not playing a game');
  }
});

$('#abort').on('click', (event) => {
  if (game.chess !== null) {
    session.send({ type: MessageType.Control, command: 0, text: 'abort' });
  } else {
    showStatusMsg('You are not playing a game');
  }
});

$('#takeback').on('click', (event) => {
  if (game.chess !== null) {
    if (game.chess.turn() === game.color) {
      pendingTakeback = 2;
      session.send({ type: MessageType.Control, command: 0, text: 'take 2'});
    } else {
      pendingTakeback = 1;
      session.send({ type: MessageType.Control, command: 0, text: 'take 1'});
    }
  } else {
    showStatusMsg('You are not playing a game');
  }
});

$('#draw').on('click', (event) => {
  if (game.chess !== null) {
    session.send({ type: MessageType.Control, command: 0, text: 'draw' });
  } else {
    showStatusMsg('You are not playing a game');
  }
});

function getGame(opponent: string, min: string, sec: string) {
  if (game.chess === null) {
    const cmd: string = (opponent !== '') ? 'match ' + opponent : 'seek';
    session.send({ type: MessageType.Control, command: 0, text: cmd + ' ' + min + ' ' + sec });
  }
}

$('#new-game').on('click', (event) => {
  if (game.chess === null) {
    session.send({ type: MessageType.Control, command: 0, text: 'getgame' });
  }
});

$('#onezero').on('click', (event) => {
  getGame(getValue('#opponent-player-name'), '1', '0');
});

$('#threezero').on('click', (event) => {
  getGame(getValue('#opponent-player-name'), '3', '0');
});

$('#threetwo').on('click', (event) => {
  getGame(getValue('#opponent-player-name'), '3', '2');
});

$('#fivezero').on('click', (event) => {
  getGame(getValue('#opponent-player-name'), '5', '0');
});

$('#fivefive').on('click', (event) => {
  getGame(getValue('#opponent-player-name'), '5', '5');
});

$('#tenfive').on('click', (event) => {
  getGame(getValue('#opponent-player-name'), '10', '5');
});

$('#fifteenzero').on('click', (event) => {
  getGame(getValue('#opponent-player-name'), '15', '0');
});

$('#custom-control').on('click', (event) => {
  if (game.chess === null) {
    const min: string = getValue('#custom-control-min');
    const sec: string = getValue('#custom-control-sec');
    getGame(getValue('#opponent-player-name'), min, sec);
  }
});

$('#sound-toggle').on('click', (event) => {
  // todo: disable sound
  const iconClass = 'dropdown-icon fa fa-volume-' + (soundToggle ? 'up' : 'off');
  $('#sound-toggle').html('<span id="sound-toggle-icon" class="' + iconClass +
    '" aria-hidden="false"></span>Sounds ' + (soundToggle ? 'ON' : 'OFF'));
  soundToggle = !soundToggle;
});

$('#disconnect').on('click', (event) => {
  session.disconnect();
});

$('#login').on('click', (event) => {
  const user: string = getValue('#login-user');
  const pass: string = getValue('#login-pass');
  if (!session) {
    session = new Session(ICSMessageHandler, user, pass);
  } else {
    if (!session.isConnected()) {
      session.connect(ICSMessageHandler, user, pass);
    }
  }
  $('#login-screen').modal('hide');
});

$('#connect-user').on('click', (event) => {
  if (session && session.isConnected()) {
    session.disconnect();
  }

  $('#login-screen').modal('show');
});

$('#connect-guest').on('click', (event) => {
  if (!session) {
    session = new Session(ICSMessageHandler);
  } else {
    if (!session.isConnected()) {
      session.connect(ICSMessageHandler);
    }
  }
});

$('#colortheme-default').on('click', (event) => {
  $('#colortheme').attr('href', 'assets/css/themes/default.css');
});

$('#colortheme-green').on('click', (event) => {
  $('#colortheme').attr('href', 'assets/css/themes/green.css');
});

$('#colortheme-yellow').on('click', (event) => {
  $('#colortheme').attr('href', 'assets/css/themes/yellow.css');
});

$(window).focus(() => {
  if (game.chess) {
    board.position(game.chess.fen());
  }
});

$(window).resize(() => {
  board.resize();
  $('.chat-text').height($('#board').height() - 40);
});
