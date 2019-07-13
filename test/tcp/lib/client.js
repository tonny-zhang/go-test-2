const {Client} = require('./Socket');

let client = new Client();
client.on(Client.EVENT_CONNECT, () => {
    for (let i = 0; i<10; i++) {
        client.send(Buffer.from(''+i), 1);
    }
});
client.connect({
    port: 6006,
    // host: '127.0.0.1'
});