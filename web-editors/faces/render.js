const width = 10;
const height = 5;
let styles = [];
const colors = ["R","G","B"];

const initGrid = (ctx,dim) => {
    for (let i = 0;i < height;i++){
        for (let b = 0;b < width;b++){
            ctx.beginPath();
            ctx.rect(b * dim, i * dim, dim, dim);
            ctx.stroke();
        }
    }
}


const drawRect = (ctx,dim,idx) => {
    if (idx.x >= width || idx.y >= height){
        return;
    }
    ctx.fillStyle = "red";
    ctx.fillRect(dim * idx.x,dim * idx.y,dim,dim);
}



var c = document.getElementById("canvas");
var rect = c.getBoundingClientRect();
const ctx = c.getContext("2d");

function printMousePos(event) {
    const pos = getPos(event,rect);
    const idx = getTileIdx(pos,20);
    drawRect(ctx,20,idx);
}

const getPos = (event,rect) => {
    const points = {};
    points.x = event.clientX - rect.left;
    points.y = event.clientY - rect.top;
    return points;
}

const getTileIdx = (pos,dim) => {
    const idx = {};
    idx.x = parseInt(pos.x / dim);
    idx.y = parseInt(pos.y / dim);
    return idx;
}
document.getElementById("canvas").style.border = "thin dotted #000";
c.addEventListener("click", printMousePos);
const controls = document.getElementById("controls")
const addStyle = document.getElementById("add-style");
addStyle.onclick = () => {
    const styleIdx = styles.length;
    const colorDiv = document.createElement("div");
    let ct = 0;
    for (const color of colors){
        const iC = document.createElement("input");
        const label = document.createElement("label");
        iC.type = "number";
        iC.defaultValue = "0";
        iC.min = "0";
        iC.max = "255";
        const tempPos = ct;
        iC.onchange = (event) => {
            styles[styleIdx][tempPos] = parseInt(event.target.value);
        }
        iC.id = color+"-input-"+styleIdx;
        label.htmlFor = color+"-input-"+styleIdx;
        label.innerText = color + ":";
        colorDiv.append(label);
        colorDiv.append(iC);
        ct++;
    }
    controls.insertBefore(colorDiv,addStyle);
    styles.push([0,0,0]);
}
initGrid(ctx,20);
