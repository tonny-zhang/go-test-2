/**
 * 对 Socket 的封装，会自动处理数据包的封包和解包，及长连接的拆包粘包问题
 */
const net = require('net')
const EventEmitter = require('events');

const EVENT_CONNECT = 'connect';
const EVENT_LISTENING = 'listening';
const EVENT_ERROR = 'error';
const EVENT_COMMAND = 'command';
const EVENT_CLOSE = 'close';
const EVENT_CLIENT_CLOSE = 'client_close';
const _EVENT_SERVER_CLOSE = 'server_close';

const DELAY_RECONNECT = 5000;
const MAX_LISTENERS = 30;
const RECONNECT_NUM = 10; // 断线重连次数

class Server extends EventEmitter{
    constructor(conf) {
        super();
        const _this = this;
        _this._conf = conf;
        _this.on('error', e => {});
        _this.started = false;

        let server = net.createServer(socket => {
            let client = new Client(socket);

            // 保证优先处理系统定义 EVENT_COMMAND 事件
            client.on(EVENT_COMMAND, (data, code) => {
                _this.emit(EVENT_COMMAND, data, code, client);
            });
            client.on(EVENT_CLOSE, () => {
                _this.emit(EVENT_CLIENT_CLOSE, client);
            });

            // 触发 conect 事件，并对外提供 Client 实例
            _this.emit(EVENT_CONNECT, client);
        });
        server.setMaxListeners(MAX_LISTENERS);
        server.on('listening', () => {
            _this.started = true;
            let info = server.address();
            // 触发 listening 事件，并提供 address 信息
            _this.emit(EVENT_LISTENING, info);
        });
        server.on('error', e => {
            _this.emit(EVENT_ERROR, e);
        });
        
        _this._server = server;
    }
    address() {
        const _this = this;
        const _conf = _this._conf;
        let info = _this._server.address();
        let ip;
        if (_conf && (ip = _conf.ipForService) ){
            info.address = ip;
        }
        return info;
    }
    start(conf) {
        conf = conf || this._conf;
        this._conf = conf;
        // 支持ipc path及端口监听方式
        this._server.listen(conf.path || conf.port, conf.host);
    }
    close() {
        const _this = this;
        if (_this.started) {
            _this._server.close(() => {
                _this.started = false;
                _this.emit(EVENT_CLOSE);
            });
        }
    }
}
const Client = (() => {
    const FLAG_MESSAGE = '||||';
    let len_flag_message = FLAG_MESSAGE.length;
    /**
     * 处理 Socket 的消息传输
     * 
     * @param {Function} callback
     */
    function _deal(callback) {
        let bf_unpack = null; // 处理粘包问题
        let bf_datareaded = null; // 处理拆包问题
        /**
         * 读取 Buffer
         * @param {Buffer} bf 
         */
        function _read(bf, isReadUnpack) {
           if (bf_unpack) {
                bf = Buffer.concat([bf_unpack, bf]);

                bf_unpack = null;
            }
            
            /**
             * 单独处理拆包问题，防止数据量大时对字符串检测造成的性能下降
             */
            if (bf_datareaded) {
                let _len_total = bf_datareaded.len;
                let _data = bf_datareaded.data;
                let _len_readed = _data.length;
                let _index_end = _len_total - _len_readed;
                let _data_readed = bf.slice(0, _index_end);
                _data = Buffer.concat([_data, _data_readed]);
                _len_readed = _data.length;

                bf = bf.slice(_index_end);
                let _code = bf_datareaded.code;
                if (_len_readed == _len_total) {
                    bf_datareaded = null;
                    callback && callback(_data, _code);
                } else {
                    bf_datareaded.data = _data;
                }
            }
            let index_start = bf.indexOf(FLAG_MESSAGE); 
            // 有消息内容并保证可以读取到code和数据长度
            if (index_start > -1 && bf.length - index_start >= len_flag_message + 6) {
                let len_total = bf.length;
                let index_read = index_start + len_flag_message;
                let code = bf.readInt16LE(index_read);
                index_read += 2;
                let len_data = bf.readInt32LE(index_read);
                index_read += 4;

                let index_end = index_read + len_data;
                let data = bf.slice(index_read, index_end);

                let len_data_readed = data.length;

                if (len_data_readed < len_data) { // 数据进行了拆包处理
                    // 对只读取到一半数据情况进行处理
                    bf_datareaded = {
                        len: len_data,
                        code: code,
                        data: data
                    };
                    // bf_unpack = data;
                    // bf_unpack = bf.slice(index_start);
                } else if (index_end <= len_total) { // 进行了粘包处理
                    callback && callback(data, code);
                    let bf_remain = bf.slice(index_end);
                    if (bf_remain.length > 0) {
                        // 防止出现 “RangeError: Maximum call stack size exceeded”
                        process.nextTick(() => {
                            _read(bf_remain, true);
                        });                        
                    }
                }
            } else {
                bf_unpack = bf;
            }
        }

        return _read;
    }

    return class Client extends EventEmitter{
        constructor(socket) {
            super();
            const _this = this;

            /**
             * 定义一个空函数，防止外部不定义 error 事件
             * https://nodejs.org/dist/latest-v6.x/docs/api/events.html#events_error_events
             */
            _this.on(EVENT_ERROR, e => {
                
            });
            if (socket) {
                socket._connected = true;
            }
            let isServer = !!!socket;
            _this.isServer = isServer;
            socket = socket || new net.Socket();
            socket.on('connect', () => {
                socket._closed = false;
                socket._connected = true;
                _this.emit(EVENT_CONNECT);
            });
            
            socket.on('data', _deal((data, code) => {
                _this.emit(EVENT_COMMAND, data, code);
            }));
            socket.on('error', e => {
                if (!socket.destroyed) {
                    socket.destroy();
                }
                _this.emit(EVENT_ERROR, e);
            });
            socket.on('close', (had_error) => {
                socket._connected = false;
                _this.emit(EVENT_CLOSE);
                // 服务端主动断开时增加断线重连机制
                if (!socket._closed) {
                    let e = new Error('close by server!');
                    e.code = 'SERVERCLOSED';
                    _this.emit(_EVENT_SERVER_CLOSE, e);
                }
            });

            _this._socket = socket;
        }

        /**
         * 得到 socket 地址和端口信息
         */
        address() {
            let socket = this._socket;
            return {
                local: {
                    address: (socket.localAddress || '').replace(/^:+ffff:/, ''),
                    port: socket.localPort
                },
                remote: {
                    address: (socket.remoteAddress || '').replace(/^:+ffff:/, ''),
                    port: socket.remotePort
                }
            }
        }

        /**
         * 得到客户端唯一标识，用于缓存
         */
        getKey() {
            const _this = this;
            let info = _this.address().remote;
            let key = info.address + '_' + info.port;
            return key;
        }

        /**
         * 主动连接，主要用于构造函数参数为空时
         */
        connect(opt) {
            opt = Object.assign({
                _tryNum: 1,
                isReConnect: true
            }, opt);
            if (opt._tryNum > RECONNECT_NUM) {
                return;
            }
            const _this = this;
            const socket = _this._socket;
            function _error(e) {
                let code = e.code;
                // 当发生错误时重连 Master
                let isServerClosed = code == 'SERVERCLOSED';
                if ((opt._reconnecting && (code == 'ECONNREFUSED' || code == 'ECONNRESET')) || isServerClosed) {
                    // 服务端断开时重置重连次数
                    if (isServerClosed) {
                        opt._tryNum = 1;
                    }
                    console.log('client can not connect server, will retry '+ (opt._tryNum) +'/'+RECONNECT_NUM+' time after '+DELAY_RECONNECT+' ms!', code, isServerClosed);
                    clearTimeout(socket._ttRetry);
                    socket._ttRetry = setTimeout(() => {
                        // 已经调用 close 方法的不再重连
                        if (!socket._connected && !socket._closed) {
                            opt._tryNum = (opt._tryNum || 0) + 1;
                            opt._reconnecting = true;
                            _this.connect(opt);
                        }
                    }, DELAY_RECONNECT);
                }
            }
            _error.__id = 'connect_error';
            if (opt.isReConnect && !opt._reconnecting) {
                function _dele(arr) {
                    if (!arr) {
                        return;
                    }
                    if (!Array.isArray(arr)) {
                        arr = [arr];
                    }
                    // 删除之前绑定的内部处理重连的error事件，防止事件绑定过多
                    for (let i = 0; i<arr.length; i++) {
                        let e = arr[i];
                        if (e.__id == _error.__id) {
                            arr.splice(i--, 1);
                        }
                    }
                }
                _dele(_this._events[EVENT_ERROR]);
                _dele(_this._events[_EVENT_SERVER_CLOSE]);
                _this.on(EVENT_ERROR, _error);
                _this.on(_EVENT_SERVER_CLOSE, _error);
            }
            if (socket) {
                let pipe_path = opt.path;
                // 支持ipc path连接方式
                socket.connect(pipe_path || opt);
            }
        }
        /**
         * 发送普通 Buffer
         * 
         * @param {Buffer} bf
         * @param {Number} code
         */
        send(bf, code) {
            let _this = this;
            let socket = _this._socket;

            if (socket.destroyed) {
                _this.emit(EVENT_ERROR, new Error('can not send anything when socket is closed!'));
                return;
            }
            let bf_data = null;
            if (bf === null || bf === undefined) {
                bf_data = Buffer.alloc(0);
            } else {
                bf_data = Buffer.from(bf);
            }
            let len_data = bf_data.length;

            let bf_send = Buffer.alloc(len_flag_message + 6 + len_data);
            bf_send.write(FLAG_MESSAGE);
            let index_write = len_flag_message;
            bf_send.writeInt16LE(code, index_write); // 写入操作 code 
            index_write += 2;
            bf_send.writeInt32LE(len_data, index_write); // 写入数据长度
            index_write += 4;
            bf_data.copy(bf_send, index_write);

            socket.write(bf_send);
        }
        /**
         * 发送命令编码
         * @param {*} code 要发送的命令编码
         */
        sendCode(code) {
            this.send(null, code);
        }

        close() {
            this._socket.end();
            this._socket.destroy();
            this._socket._closed = true; // 标识已经调用了close方法
        }
        isEnabled() {
            const _socket = this._socket;
            return _socket && _socket._connected;
        }
    }
})();

Server.EVENT_LISTENING = EVENT_LISTENING;
Server.EVENT_CONNECT = EVENT_CONNECT;
Server.EVENT_ERROR = EVENT_ERROR;
Server.EVENT_CLOSE = EVENT_CLOSE;
Server.EVENT_COMMAND = EVENT_COMMAND;
Server.EVENT_CLIENT_CLOSE = EVENT_CLIENT_CLOSE;

Client.EVENT_CONNECT = EVENT_CONNECT;
Client.EVENT_ERROR = EVENT_ERROR;
Client.EVENT_CLOSE = EVENT_CLOSE;
Client.EVENT_COMMAND = EVENT_COMMAND;

module.exports.Server = Server;
module.exports.Client = Client;