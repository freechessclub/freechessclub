// Copyright 2017 Free Chess Club.

export class History {
  private board: any;
  private moves: string[];
  private id: number;

  constructor(board: any, initialPosition: string) {
    this.board = board;
    this.moves = [ initialPosition ];
    this.id = 0;

    $('#move-history').empty();

    $('#fast-backward').off('click');
    $('#fast-backward').on('click', () => this.beginning());
    $('#backward').off('click');
    $('#backward').on('click', () => this.backward());
    $('#forward').off('click');
    $('#forward').on('click', () => this.forward());
    $('#fast-forward').off('click');
    $('#fast-forward').on('click', () => this.end());

    (window as any).showMove = (id: number) => {
      if (this) {
        this.display(id);
      }
    };

    $('#collapse-history').on('hidden.bs.collapse', () => {
      $('#history-toggle-icon').removeClass('fa-toggle-up').addClass('fa-toggle-down');
    });
    $('#collapse-history').on('shown.bs.collapse', () => {
      $('#history-toggle-icon').removeClass('fa-toggle-down').addClass('fa-toggle-up');
    });
    if ($(window).width() < 767) {
      $('#collapse-history').collapse('hide');
    }
  }

  public add(move: any, fen: string): void {
    this.moves.push(fen);
    this.id = this.moves.length - 1;
    this.update(move);
  }

  public length(): number {
    return this.moves.length - 1;
  }

  public display(id?: number): void {
    if (id !== undefined) {
      this.id = id;
    }
    if (this.id >= 0 && this.id < this.moves.length) {
      this.board.position(this.moves[this.id]);
    }
  }

  public beginning(): void {
    this.display(0);
  }

  public backward(): void {
    if (this.id > 0) {
      this.display(this.id - 1);
    }
  }

  public forward(): void {
    if (this.id < this.moves.length - 1) {
      this.display(this.id + 1);
    }
  }

  public end(): void {
    this.display(this.moves.length - 1);
  }

  public undo(): void {
    if (this.id > 0) {
      this.display(this.id - 1);
      this.moves.pop();
    }
  }

  private update(move: any): void {
    const id: number = this.length();
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
}

export default History;
