var chat = new ReconnectingWebSocket(location.protocol.replace("http","ws") + "//" + location.host + "/ws");
var connected = false;
var myHandle = "";

var game = new Chess();

var tabs;
var msgType = {
  ctl: 0,
  chTell: 1,
  pTell: 2,
  gameMove: 3,
  unknown: 4
};

var highlightSquare = function(square) {
  var squareEl = $('#board .square-' + square);
  var background = '#e6ffdd';
  if (squareEl.hasClass('black-3c85d') === true) {
    background = '#278881';
  }
  squareEl.css('background', background);
};

var unHighlightSquare = function(square) {
  if (square !== undefined) {
    $('#board .square-' + square).css('background', '');
  } else {
    $('#board .square-55d63').css('background', '');
  }
}

var validMoves;

var onDragStart = function(source, piece, position, orientation) {
  if (game.game_over() === true ||
      (game.turn() === 'w' && piece.search(/^b/) !== -1) ||
      (game.turn() === 'b' && piece.search(/^w/) !== -1)) {
    return false;
  }

  // get list of possible moves for this square
  validMoves = game.moves({
    square: source,
    verbose: true
  });

  if (validMoves.length === 0) return;
  highlightSquare(source);
  for (var i = 0; i < validMoves.length; i++) {
    highlightSquare(validMoves[i].to);
  }
};

var onDrop = function(source, target) {
  // see if the move is legal
  var move = game.move({
    from: source,
    to: target,
    promotion: 'q' // TODO: Allow non-queen promotes
  });

  // illegal move
  if (move === null) {
    unHighlightSquare(source);
    if (validMoves.length > 0) {
      for (var i = 0; i < validMoves.length; i++) {
        unHighlightSquare(validMoves[i].to);
      }
      validMoves = [];
    }
    return 'snapback';
  }

  // valid move
  unHighlightSquare();
  highlightSquare(source);
  highlightSquare(target);
};

// update the board position after the piece snap
// for castling, en passant, pawn promotion
var onSnapEnd = function() {
  board.position(game.fen());
};

var cfg = {
  position: 'start',
  showNotation: true,
  draggable: true,
  onDragStart: onDragStart,
  onDrop: onDrop,
  onSnapEnd: onSnapEnd
};
var board = ChessBoard('board', cfg);

// enable tooltips
$(function () {
  $('[data-toggle="tooltip"]').tooltip()
})

// Allow chat card to be collapsed
$('#collapse-chat').on('hidden.bs.collapse', function () {
  $("#chat-toggle").text("+");
})
$('#collapse-chat').on('show.bs.collapse', function () {
  $("#chat-toggle").text(String.fromCharCode(8211));
})

jQuery(document.body).on('click', '.closeTab', function(event) {
  var tabContentId = $(this).parent().attr("href");
  $(this).parent().remove();
  delete tabs["content-"+tabContentId.substr(1)];
  $('#tabs a:last').tab('show');
  $(tabContentId).remove();
});

$(document).on('shown.bs.tab', 'a[data-toggle="tab"]', function (e) {
  var tab = $(e.target);
  tab.css('color', 'black');
});

function handleChatMsg(tab, data) {
  who = "";
  var tabheader = $("#" + $("ul#tabs a.active").attr("id"));
  if (data.hasOwnProperty('handle')) {
    var textclass = "";
    if (myHandle == data.handle) {
      textclass = " class=\"mine\"";
    }
    who = "<strong"+textclass+">"+$('<span/>').text(data.handle).html()+"</strong>: ";
    if (data.type == msgType.chTell) {
      tabheader = $("#"+data.channel);
    } else {
      tabheader = $("#"+data.handle);
    }
  }
  tab.append(who +
    $('<span/>').text(data.text).html()+"</br>");

  if (tabheader.hasClass('active')) {
    tab.scrollTop(tab[0].scrollHeight);
  } else {
    tabheader.css('color', 'red');
  }
}

chat.onmessage = function(message) {
  var data = JSON.parse(message.data);
  switch(data.type) {
    case msgType.ctl:
      if (connected == false && data.command == 1) {
        connected = true;
        myHandle = data.text;
        $("#chat-status").text("Connected as " + myHandle);
      }
      break;
    case msgType.chTell:
      var tab = tabs[data.channel];
      handleChatMsg(tab, data)
      break;
    case msgType.pTell:
      var tab;
      if (!tabs.hasOwnProperty(data.handle)) {
        $('<a class="flex-sm-fill text-sm-center nav-link" data-toggle="tab" href="#content-'+data.handle+'" id="'+data.handle+'" role="tab">'+data.handle+'<span class="btn btn-default btn-sm closeTab">×</span></a>').appendTo('#tabs');
        $('<div class="tab-pane chat-text" id="content-'+data.handle+'" role="tabpanel"></div>').appendTo('.tab-content');
        $(".chat-text").height($("#board").height()-115);
        tab = $("#content-"+data.handle);
        tabs[data.handle] = tab;
      } else {
        tab = tabs[data.handle];
      }
      handleChatMsg(tab, data);
      break;
    case msgType.gameMove:
      board.position(data.fen, false);
      break;
    case msgType.unknown:
    default:
      var tab = $("ul#tabs a.active").attr("id");
      handleChatMsg(tabs[tab], data);
      break;
  }
};

chat.onclose = function(){
  connected = false;
  $("#chat-status").text("Disconnected");
};

$("#input-form").on("submit", function(event) {
  event.preventDefault();
  var tab = $("ul#tabs a.active").attr("id");
  var text;
  if (!$("#input-command").is(':checked')) {
    if ($("#input-text").val().charAt(0) != "@") {
      msg = $("#input-text").val();
      text = "t " + tab + " " + msg;
      handleChatMsg(tabs[tab], { type: msgType.chTell, channel: tab, handle: myHandle, text: msg });
    } else {
      text = $("#input-text").val().substr(1);
    }
  } else {
    if ($("#input-text").val().charAt(0) != "@") {
      text = $("#input-text").val();
    } else {
      text = $("#input-text").val().substr(1);
    }
  }
  chat.send(JSON.stringify({ type: msgType.pTell, handle: myHandle, text: text }));
  $("#input-text").val("");
});

$(document).ready(function() {
  connected = false;
  $("#chat-status").text("Connecting...");
  $(".chat-text").height($("#board").height()-115);
  tabs = { "53": $("#content-53") };
});

$(window).resize(function() {
  board.resize();
  $(".chat-text").height($("#board").height()-115);
});
