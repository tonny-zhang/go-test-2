require('child_process').exec([
    'main.exe',
    'C:\\Users\\Administrator\\Desktop\\测试图片\\8098773_125154013000_2.jpg',
    '1.jpg'
].join(' '), (...args) => {
    console.log(args)
});
