const express = require('express')
const app = express()

app.get('/', function (req, res) {
    const helloTo = process.env.HELLO_TO || 'world'
    const appName = process.env.APP_NAME || 'some app'
    const appRuntime = process.env.APP_RUNTIME || 'some runtime'
    const podName = process.env.POD_NAME || 'some pod'
    res.send(`Hello ${helloTo}! This is app ${appName}, using runtime ${appRuntime} in pod ${podName}`)
})

app.listen(3000, function () {
    console.log('Example app listening on port 3000!')
})