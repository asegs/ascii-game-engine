const drawTiles = (ctx,n,dim) => {
    let toggle = false;
    for (let i = 0;i < n;i++){
        for (let b = 0;b < n;b++){
            if (toggle){
                ctx.fillStyle = "red";
                ctx.fillRect(b * dim, i * dim, dim , dim);
            }else{
                ctx.fillStyle = "green";
                ctx.fillRect(b * dim, i * dim, dim , dim);
            }
            toggle = !toggle;
        }
    }
}


const drawRect = (ctx,dim,idx) => {
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
c.addEventListener("click", printMousePos);
drawTiles(ctx,5,20);
