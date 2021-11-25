const width = 20;
const height = 10;
const xDim = 20;
const yDim = 40;
let styles = [];
let selected = 0;
const pixels = new Array(height);
let x = 0;
let y = 0;
let toggled = false;

const initGrid = (ctx) => {
    for (let i = 0;i < height;i++){
        for (let b = 0;b < width;b++){
            ctx.beginPath();
            ctx.rect(b * xDim, i * yDim, xDim, yDim);
            ctx.closePath();
            ctx.stroke();
        }
    }
}

const highlightBox = (idx,undo) => {
    ctx.beginPath();
    if (undo){
        ctx.lineWidth = "2";
        const color = pixels[idx.y][idx.x] !== -1 ? styles[pixels[idx.y][idx.x]] : "white";
        ctx.beginPath();
        ctx.rect(idx.x * xDim + 3,idx.y * yDim + 3,xDim - 6,yDim - 6);
        ctx.closePath();
        ctx.stroke();
        drawRawRect(ctx,{x:x,y:y},color);
        ctx.strokeStyle = "black";
        ctx.beginPath();
        ctx.rect(idx.x * xDim,idx.y * yDim,xDim,yDim);
        ctx.closePath();
        ctx.stroke();
    }else {
        ctx.lineWidth = "2";
        ctx.strokeStyle = "yellow";
        ctx.beginPath();
        ctx.rect(idx.x * xDim + 3,idx.y * yDim + 3,xDim - 6,yDim - 6);
        ctx.closePath();
        ctx.stroke();
    }
}

const drawRect = (ctx,idx, allowUndo) => {
    if (styles.length === 0){
        document.getElementById("add-style").focus();
        return;
    }
    const color = styles[selected];
    if (selected === pixels[idx.y][idx.x] && allowUndo){
        pixels[idx.y][idx.x] = -1;
        ctx.fillStyle = "#FFF";

    }else{
        ctx.fillStyle = color;
        pixels[idx.y][idx.x] = selected;
    }
    ctx.fillRect(xDim * idx.x,yDim * idx.y,xDim,yDim);
}

const drawRawRect = (ctx,idx,color) => {
    ctx.fillStyle = color;
    ctx.fillRect(idx.x * xDim,idx.y * yDim,xDim,yDim);
}



var c = document.getElementById("canvas");
var rect = c.getBoundingClientRect();
const ctx = c.getContext("2d");

function handleClick(event) {
    const pos = getPos(event,rect);
    const idx = getTileIdx(pos);
    if ( idx.x >= width || idx.y >= height){
        return;
    }
    highlightBox({x:x,y:y},true);
    x = idx.x;
    y = idx.y;
    highlightBox(idx,false)
    drawRect(ctx,idx,true);
}

const getPos = (event,rect) => {
    const points = {};
    points.x = event.clientX - rect.left;
    points.y = event.clientY - rect.top;
    return points;
}

const getTileIdx = (pos) => {
    const idx = {};
    idx.x = Math.trunc(pos.x / xDim);
    idx.y = Math.trunc(pos.y / yDim);
    return idx;
}

const downloadString = (text) => {
    const link = document.createElement('a');
    link.download = 'face.txt';
    const blob = new Blob([text], {type: 'text/plain'});
    link.href = window.URL.createObjectURL(blob);
    link.click();
}
document.getElementById("canvas").style.border = "thin dotted #000";
c.addEventListener("click", handleClick);
const controls = document.getElementById("controls")
const addStyle = document.getElementById("add-style");
addStyle.onclick = () => {
    const styleIdx = styles.length;
    const picker = document.createElement("input");
    picker.type = "color";
    picker.onclick = () => {
        selected = styleIdx;
    }
    picker.onchange = (event) => {
        styles[styleIdx] = event.target.value;
        for (let i = 0;i<height;i++){
            for (let b = 0;b<width;b++){
                if (pixels[i][b] === styleIdx){
                    drawRect(ctx,{x:b,y:i},false);
                }
            }
        }
    }

    styles.push("#000");
    controls.insertBefore(picker,addStyle);
    controls.insertBefore(document.createElement("br"),addStyle);
}
initGrid(ctx);
const saveFace = document.getElementById("save");
saveFace.onclick = () => {
    const body = new Array(width * height + height);
    let current = 0;
    for (let i = 0;i<height;i++){
        for (let b = 0;b<width;b++){
            body[current] = pixels[i][b];
            current++;
        }
        body[current] = "\n";
        current++;
    }
    downloadString(body.join(","));
}
for (let i = 0 ; i < height ; i++ ){
    pixels[i] = new Array(width).fill(-1);
}

const handleMovement = (key) => {
    highlightBox({x:x,y:y},true);
    switch (key){
        case "ArrowRight":
            x ++;
            break;
        case "ArrowLeft":
            x --;
            break;
        case "ArrowUp":
            y --;
            break;
        case "ArrowDown":
            y ++;
            break;
    }
    if (x >= width){
        x = 0;
    }else if (x < 0){
        x = width - 1;
    }
    if (y >= height){
        y = 0;
    }else if (y < 0){
        y = height - 1;
    }

    highlightBox({x:x,y:y},false);
    if (toggled){
        drawRect(ctx,{x:x,y:y},false);
    }

}

window.addEventListener("keydown",(event) => {
    const key = event.key;
    if (key.substr(0,5) === "Arrow"){
        handleMovement(key);
    }else if (key === " "){
        drawRect(ctx,{x:x,y:y},true);
    }
})

highlightBox({x:x,y:y},false);

const check = document.getElementById("draw-on-move");
check.onchange = () => {
    toggled = !toggled;
}