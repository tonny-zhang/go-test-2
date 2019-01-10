var data = require('./data/NoviceTrain.json');
let $ = require('./lib/j')

$('.wrapper').on('click', (e) => {
    console.log(e.offsetX, e.offsetY);
});
let _isHaveValGlobal;
function _writeBf(width, height) {
    let canvas3 = $('<canvas width="'+width+'" height="'+height+'" style="z-index: 100"/>').appendTo('.wrapper').get(0);
    let cxt3 = canvas3.getContext('2d');
    cxt3.fillStyle = 'red';
    let bfData = require('fs').readFileSync('E:/source/go/src/test/parseMesh/meshData/NoviceTrain.dat');

    let _width = bfData.readInt32LE();
    let _height = bfData.readInt32LE(4);
    console.log('_width = '+ _width+', _height = '+_height+', width = '+width+', height = '+height);
    cxt3.clearRect(0, 0, width, height);
    // cxt3.fillStyle = 'rgba(0, 0, 0, 1)';
    // cxt3.fillRect(0, 0, width, height);
    let imgData = cxt3.getImageData(0, 0, width, height);
    let data = imgData.data;
    for (let i = 0; i<width; i++) {
        for (let j = 0; j<height; j++) {
            // let indexBf = i + j * width;
            let indexPixel = i * height + j;
            let indexBf = Math.floor(indexPixel/8) + 8;
            let indexBit = indexPixel % 8;
            let val = bfData[indexBf];
            val = val >> (7 - indexBit) & 1;
            if (val) {
                let indexPixel = 4 * (i + width * j);
                data[indexPixel] = 0;
                data[indexPixel + 1] = 255;
                data[indexPixel + 2] = 0;
                data[indexPixel + 3] = 100;
            }
        }
    }
    cxt3.putImageData(imgData, 0, 0);
    _isHaveValGlobal = function(x, y) {
        let indexPixel = x * height + y;
        let indexBf = Math.floor(indexPixel/8);
        let indexBit = indexPixel % 8;
        let val = bfData[indexBf];
        val = val >> (7 - indexBit) & 1;
        return !!val;
    }
}
var img = new Image();
img.onload = function() {
    let {width, height} = this;

    _writeBf(width, height);
    $('<img src="'+this.src+'" style="width='+width+'px; height:'+height+'px;"/>').appendTo('.wrapper');
    let canvas = $('<canvas width="'+width+'" height="'+height+'"/>').appendTo('.wrapper').get(0);
    let cxt = canvas.getContext('2d');
    let canvas2 = $('<canvas width="'+width+'" height="'+height+'"/>').appendTo('.wrapper').get(0);
    let cxt2 = canvas2.getContext('2d');
    cxt2.fillStyle = 'red';
    cxt.fillStyle = 'rgba(0, 0, 0, 0)';
    cxt.fillRect(0, 0, width, height);
    cxt.fillStyle = 'rgba(255, 255, 255, 0.3)';

    let triangles = data.triangles;
    let vertices = data.vertices;
    function _getPoint(index) {
        let v = vertices[index];
        // let vNew = [v[0] * 100 + width/2, v[1] * 100 + height/2];
        let vNew = [v[0] * 100 + width/2, height/2-v[1] * 100];
        return vNew;
    }
    return;
    for (var i = 0, j = triangles.length; i<j; i += 3){
        let one = _getPoint(triangles[i]);
        let two = _getPoint(triangles[i+1]);
        let three = _getPoint(triangles[i+2]);
        cxt.beginPath();
        cxt.moveTo(...one);
        cxt.lineTo(...two);
        cxt.lineTo(...three);
        cxt.fill();
    }
    
    let imageData = cxt.getImageData(0, 0, width, height).data;
    function _isHaveVal(x, y) {
        if (_isHaveValGlobal) {
            return _isHaveValGlobal(x, y);
        }
        let index = 4 * (x + width * y);
        let r = imageData[index + 0];
        let g = imageData[index + 1];
        let b = imageData[index + 2];
        let a = imageData[index + 3];
        // console.log('index = ', index, r, g, b, a);
        return a !== 0;
    }
    function _getColorRnd() {
        function _getN() {
            return Math.floor(255 * Math.random());
        }
        return 'rgba('+[_getN(), _getN(), _getN(), 1].join()+')'
    }
    // console.log(_isHaveVal(95, 43), imageData[0], imageData[1], imageData[2], imageData[3]);

    var points = [];
    
    for (let i = 0; i<100000; i++) {
        points.push([Math.floor(Math.random() * width), Math.floor(Math.random() * height)]);
    }
    const STEP = 1; 
    function run() {
        let start = new Date();
        cxt2.clearRect(0, 0, width, height);
        for (let i = 0, j = points.length; i<j; i++) {
            let p = points[i];
            let x_current = p[0];
            let y_current = p[1];
            let x_next = x_current + (Math.random() < 0.5? STEP: -STEP);
            let y_next = y_current + (Math.random() < 0.5? STEP: -STEP);
            if (_isHaveVal(x_next, y_next)) {
                cxt2.fillStyle = _getColorRnd();
                cxt2.fillRect(x_next, y_next, 2, 2);
                points[i] = [x_next, y_next];
            }
        }
        console.log(new Date() - start);
        setTimeout(run, 30);
    }
    run();
}
img.src = './data/NoviceTrain.jpg';