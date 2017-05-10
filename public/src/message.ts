// Copyright 2017 The Free Chess Club.

// Type of messages we receive from the proxy
export enum MessageType {
  Control = 0,
  ChannelTell,
  PrivateTell,
  GameMove,
  GameStart,
  GameEnd,
  Unknown,
}

export default MessageType;
