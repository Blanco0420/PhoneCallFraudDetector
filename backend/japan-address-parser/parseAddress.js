const japa = require('jp-address-parser');


const processAddress = async (address) =>{
  try {
    const address = process.argv[2];
    var result = { }
    result = { ...result, data: await japa.parse(address) }
    console.log(JSON.stringify(result))
  }
  catch(e){
    console.log(JSON.stringify({ errorString: e.code, success: false }))
  }
  /*
  { prefecture: '東京都',
    city: '北区',
    town: '東十条',
    chome: 6,
    ban: 28,
    go: 70,
    left: '' }
  */
  // console.log(await japa.normalize('京都府京都市東山区本町22-489-1'))
  // 京都府京都市東山区本町二十二丁目４８９番１号
}


  // const address = process.argv[2];
  const address = "神奈川県横浜市西区高島2514リバース横浜403"
  processAddress(address)