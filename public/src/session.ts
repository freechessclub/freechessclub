// Copyright 2017 The Free Chess Club.

import MessageType from './message';

export default class Session {
  private connected: boolean;
  private handle: string;
  private websocket: WebSocket;

  constructor(onMessage: (msg: any) => void, user?: string, pass?: string) {
    this.connected = false;
    this.handle = '';
    this.connect(onMessage, user, pass);
  }

  public getHandle(): string {
    return this.handle;
  }

  public setHandle(handle: string): void {
    this.connected = true;
    this.handle = handle;
    $('#chat-status').text('Connected as ' + handle);
  }

  public isConnected(): boolean {
    return this.connected;
  }

  public connect(onMessage: (msg: any) => void, user?: string, pass?: string) {
    $('#chat-status').text('Connecting...');
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

    this.websocket = new WebSocket(
      location.protocol.replace('http', 'ws') + '//' + location.host + '/ws' + loginOptions);
    this.websocket.onmessage = onMessage;
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
      $('#chat-status').text('Disconnecting...');
      this.websocket.close();
      this.connected = false;
      this.handle = '';
    }
  }

  public reset(evt) {
    $('#chat-status').text('Disconnected');
  }

  public send(payload: string | object) {
    if (!this.isConnected()) {
      throw new Error('Session not connected.');
    }

    let data: string;
    if (typeof payload === 'object') {
      data = JSON.stringify(payload);
    } else {
      data = payload;
    }
    this.websocket.send(data);
  }
}
