// Copyright 2017 The Free Chess Club.

export default class History {
  private board: any;
  private moves: string[];
  private id: number;

  constructor(board: any, initialPosition: string) {
    this.board = board;
    this.moves = [ initialPosition ];
    this.id = 0;
  }

  public add(move: string): void {
    this.moves.push(move);
    this.id = this.moves.length - 1;
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
}
