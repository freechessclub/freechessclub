// Copyright 2017 Free Chess Club.

export const enum MessageType {
  Control = 0,
  ChannelTell,
  PrivateTell,
  GameMove,
  GameStart,
  GameEnd,
  Unknown,
}

export class Session {
  private connected: boolean;
  private handle: string;
  private websocket: WebSocket;
  private onRecv: (msg: any) => void;

  constructor(onRecv: (msg: any) => void, user?: string, pass?: string) {
    this.connected = false;
    this.handle = '';
    this.onRecv = onRecv;
    this.connect(user, pass);
  }

  public getHandle(): string {
    return this.handle;
  }

  public setHandle(handle: string): void {
    this.connected = true;
    this.handle = handle;
    $('#chat-status').html('<span class="badge badge-success">Connected</span><span class="align-middle"> '
      + handle + '</span>');
  }

  public isConnected(): boolean {
    return this.connected;
  }

  public connect(user?: string, pass?: string) {
    $('#chat-status').html('<span class="badge badge-info">Connecting...</span>');
    const login = (user !== undefined && pass !== undefined);
    let loginOptions = '';
    if (login) {
      loginOptions += '?login=1';
    }

    let text = '[' + user;
    if (pass !== undefined && pass.length > 0) {
      text += ',' + btoa(pass);
    }
    text += ']';

    let host = location.host;
    if (host === '') {
      host = 'www.freechess.club';
    }

    let protocol = 'ws://';
    if (location.protocol === 'https:' || location.protocol === 'file:') {
      protocol = 'wss://';
    }

    this.websocket = new WebSocket(protocol + host + '/ws' + loginOptions);
    this.websocket.onmessage = (message: any) => {
      const data = JSON.parse(message.data);
      if (Array.isArray(data)) {
        data.map((m) => this.onRecv(m));
      } else {
        this.onRecv(data);
      }
    };
    this.websocket.onclose = this.reset;
    if (login) {
      this.websocket.onopen = () => {
        this.websocket.send(
          JSON.stringify({ type: MessageType.Control, command: 1, text }));
      };
    }
  }

  public disconnect() {
    if (this.isConnected()) {
      $('#chat-status').html('<span class="badge badge-info">Disconnecting...</span>');
      this.websocket.close();
      this.connected = false;
      this.handle = '';
    }
  }

  public reset(evt) {
    $('#chat-status').html('<span class="badge badge-danger">Disconnected</span>');
  }

  public send(payload: string | object) {
    if (!this.isConnected()) {
      throw new Error('Session not connected.');
    }
    this.websocket.send(JSON.stringify(payload));
  }
}

export default Session;
