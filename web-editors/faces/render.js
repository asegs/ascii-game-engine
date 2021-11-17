const width = 10;
const height = 10;
let styles = [];
const colors = ["R","G","B"];
let selected = 0;

const RGBToHex= (r,g,b)=> {
    r = r.toString(16);
    g = g.toString(16);
    b = b.toString(16);

    if (r.length === 1)
        r = "0" + r;
    if (g.length === 1)
        g = "0" + g;
    if (b.length === 1)
        b = "0" + b;

    return "#" + r + g + b;
}

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
    const color = styles[selected];
    ctx.fillStyle = RGBToHex(color[0],color[1],color[2]);
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
    colorDiv.id = "color-div-" + styleIdx;
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
            const colorSlice = styles[styleIdx];
            styles[styleIdx][tempPos] = parseInt(event.target.value);
            const cBtn = document.getElementById("color-btn-"+styleIdx);
            if (cBtn){
             cBtn.style.backgroundColor = RGBToHex(colorSlice[0],colorSlice[1],colorSlice[2]);
            }
        }
        iC.id = color+"-input-"+styleIdx;
        label.htmlFor = color+"-input-"+styleIdx;
        label.innerText = color + ":";
        colorDiv.append(label);
        colorDiv.append(iC);
        ct++;
    }
    const colorBtn = document.createElement("button");
    styles.push([0,0,0]);
    colorBtn.style.backgroundColor = "#000";
    colorBtn.id = "color-btn-" + styleIdx;
    colorBtn.innerText = "EXAMPLE";
    colorDiv.append(colorBtn);
    colorDiv.onclick = () => {
        selected = styleIdx;
        for (let i = 0;i<styles.length;i++){
            if (selected === i){
                document.getElementById("color-div-" + i).style.backgroundColor = "#c2fcfc";
            } else {
                document.getElementById("color-div-" + i).style.backgroundColor = "#FFF";
            }
        }

    }
    controls.insertBefore(colorDiv,addStyle);
}
initGrid(ctx,20);
