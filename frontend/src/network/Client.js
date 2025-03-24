import { decodeServerMessage, encodeServerMessage } from './message.pb.js';

export default class Client {
  /**
   * @param {string} url - The WebSocket server URL.
   * @param {object} [options] - Optional configuration.
   * @param {number} [options.reconnectInterval=3000] - Time in milliseconds between reconnection attempts.
   */
  constructor(url, options = {}) {
    this.url = url;
    this.ws = null;
    this.reconnectInterval = options.reconnectInterval || 3000;

    // Event callbacks; assign these as needed:
    this.onOpen = null;
    this.onClose = null;
    this.onError = null;
    this.onMessage = null;
  }

  /**
   * Establishes the WebSocket connection and sets up event handlers.
   */
  connect() {
    this.ws = new WebSocket(this.url);
    this.ws.binaryType = 'arraybuffer';

    this.ws.onopen = () => {
      console.log('WebSocket connected.');
      if (typeof this.onOpen === 'function') {
        this.onOpen();
      }
    };

    this.ws.onmessage = (event) => {
      try {
        // Convert the received ArrayBuffer into a Uint8Array.
        const data = new Uint8Array(event.data);
        // Decode the message using the generated protobuf function.
        const message = decodeServerMessage(data);
        if (typeof this.onMessage === 'function') {
          this.onMessage(message);
        }
      } catch (err) {
        console.error('Error decoding server message:', err);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      if (typeof this.onError === 'function') {
        this.onError(error);
      }
    };

    this.ws.onclose = () => {
      console.log('WebSocket connection closed.');
      if (typeof this.onClose === 'function') {
        this.onClose();
      }
      // Attempt reconnection after the specified interval.
      setTimeout(() => this.connect(), this.reconnectInterval);
    };
  }

  /**
   * Sends a message over the WebSocket.
   * @param {object} messageObj - A message object following your ServerMessage schema.
   */
  send(messageObj) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      try {
        // Encode the message object into binary using your generated function.
        const encoded = encodeServerMessage(messageObj);
        this.ws.send(encoded);
      } catch (err) {
        console.error('Error encoding or sending message:', err);
      }
    } else {
      console.warn('WebSocket is not open. Cannot send message.');
    }
  }

  /**
   * Closes the WebSocket connection.
   */
  disconnect() {
    if (this.ws) {
      this.ws.close();
    }
  }
}

