//The width (horizontal number of tiles) in the editor
const width = 20;
//The height (vertical number of tiles) in the editor
const height = 10;
//The horizontal width of each tile
const xDim = 20;
//The vertical height of each tile
const yDim = 40;
//A list of all the colors currently in use, will change when some are added
let styles = [];
//The index of the currently selected color from styles
let selected = 0;
//A list of all the tiles, imagine a grid
const pixels = new Array(height);
//The current x position of the selected tile
let x = 0;
//The current y position of the selected tile
let y = 0;
//Controls if moving the selected tile draws that color (try it out!)
let toggled = false;

/*
Draws the initial black lined grid of tiles on a canvas (ctx)
 */
const initGrid = (ctx) => {
    //for each row vertically
    for (let i = 0;i < height;i++){
        //for each cell in each row
        for (let b = 0;b < width;b++){
            //Start a path, go in a rectangle of the defined size, and then paint that rectangle black
            ctx.beginPath();
            ctx.rect(b * xDim, i * yDim, xDim, yDim);
            ctx.closePath();
            ctx.stroke();
        }
    }
}

/*
Puts the yellow glow around a box
 */
const highlightBox = (idx,undo) => {
    ctx.beginPath();
    ctx.lineWidth = "2";
    //If we are taking away the yellow glow
    if (undo){
        //Draw a rectangle of the original color over the tile
        ctx.strokeStyle = pixels[idx.y][idx.x] !== -1 ? styles[pixels[idx.y][idx.x]] : "white";
        ctx.rect(idx.x * xDim + 3,idx.y * yDim + 3,xDim - 6,yDim - 6);
        ctx.closePath();
        ctx.stroke();
        //If the original color was white, put the black edge around it
        if (ctx.strokeStyle === "white"){
            ctx.strokeStyle = "black";
            ctx.beginPath();
            ctx.rect(idx.x * xDim,idx.y * yDim,xDim,yDim);
            ctx.closePath();
            ctx.stroke();
        }
    }else {
        //Draw a yellow highlight over the tile
        ctx.strokeStyle = "yellow";
        ctx.rect(idx.x * xDim + 3,idx.y * yDim + 3,xDim - 6,yDim - 6);
        ctx.closePath();
        ctx.stroke();
    }
}


/*
Fills in a rectangle of a certain color
 */
const drawRect = (ctx,idx, allowUndo) => {
    //If there are no styles yet, jump to the "add new styles" button and then stop
    if (styles.length === 0){
        document.getElementById("add-style").focus();
        return;
    }
    //Set "color" equal to the current color from styles
    const color = styles[selected];
    //If the selected color is the same as the color already there, make it white again
    if (selected === pixels[idx.y][idx.x] && allowUndo){
        pixels[idx.y][idx.x] = -1;
        ctx.fillStyle = "#FFFFFF";

    }else{
        //If the selected color is different from what is there, paint that tile the selected color
        ctx.fillStyle = color;
        pixels[idx.y][idx.x] = selected;
    }
    ctx.fillRect(xDim * idx.x,yDim * idx.y,xDim,yDim);
}
//"c" is the variable for the canvas, where things are being drawn.
const c = document.getElementById("canvas");
//"rect" is the actual size of the canvas.
const rect = c.getBoundingClientRect();
//"ctx" is the 2d part of the canvas we're going to write on.
const ctx = c.getContext("2d");

/*
When the mouse is clicked, figure out what tile it was on and highlight that box, then draw in it
 */
function handleClick(event) {
    //Get the coordinates of the click
    const pos = getPos(event,rect);
    //Figure out what tile is at those coordinates
    const idx = getTileIdx(pos);
    //If the click is off the grid, ignore it
    if ( idx.x >= width || idx.y >= height){
        return;
    }
    //Remove the highlight from wherever it was before
    highlightBox({x:x,y:y},true);
    //Set the new location to be where the new tile is
    x = idx.x;
    y = idx.y;
    //Highlight the new tile yellow
    highlightBox(idx,false)
    //Shade in the new tile with the selected color
    drawRect(ctx,idx,true);
}

/*
Gets the coordinates on the canvas of where the mouse was clicked
 */
const getPos = (event,rect) => {
    const points = {};
    //The x position of the click on the screen, minus the x position of the left edge of the canvas
    points.x = event.clientX - rect.left;
    //The y position of the click on the screen, minus the y position of the top edge of the canvas
    points.y = event.clientY - rect.top;
    return points;
}

/*
Figures out what tile is at a certain pair of coordinates
 */
const getTileIdx = (pos) => {
    const idx = {};
    //The x position of the tile is the distance from the left, divided by the width of one tile
    idx.x = Math.trunc(pos.x / xDim);
    //The y position of the tile is the distance from the top, divided by the height of one tile
    idx.y = Math.trunc(pos.y / yDim);
    return idx;
}

/*
Saves a string of text as a file
 */
const downloadString = (text) => {
    //Creates 'a' element, or link
    const link = document.createElement('a');
    //Sets the name of the file at that link to "face.txt"
    link.download = 'face.txt';
    //Creates the file to be downloaded, with the type "text/plain", normal text
    const blob = new Blob([text], {type: 'text/plain'});
    //Makes the object at the link be that file
    link.href = window.URL.createObjectURL(blob);
    //Simulates the user clicking on the link to that file
    link.click();
}

//Sets the border of the tile canvas to be black, thin, and dotted
document.getElementById("canvas").style.border = "thin dotted #000";
//Calls the "handleClick" function when someone clicks on the tile canvas
c.addEventListener("click", handleClick);

//Grabs the object which is the control panel on the right
const controls = document.getElementById("controls");
//Grabs the button which has a "+" sign and adds a style
const addStyle = document.getElementById("add-style");

/*
Defines a function to do when someone clicks "+" to add a style
 */
addStyle.onclick = () => {
    //Gets the next number for the styles list, if there are 2 elements in it, gets the 2nd index
    const styleIdx = styles.length;
    //Creates an input field
    const picker = document.createElement("input");
    //Sets the input field type to be a color picker
    picker.type = "color";
    //Says that when the color picker is clicked on, we set the current color to whatever it was at
    picker.onclick = () => {
        selected = styleIdx;
    }
    /*
    Describes what happens when the user changes the color in the color picker
     */
    picker.onchange = (event) => {
        //Sets the color in styles to be the new color
        styles[styleIdx] = event.target.value;
        //Looks through all cells
        for (let i = 0;i<height;i++){
            for (let b = 0;b<width;b++){
                //If a cell was already the original color, changes it to the new color
                if (pixels[i][b] === styleIdx){
                    //Does that by drawing a rectangle of the new color over it
                    drawRect(ctx,{x:b,y:i},false);
                }
            }
        }
    }
    //Adds a new style to the styles list, default black
    styles.push("#000000");
    //Adds a new color picker to controls
    controls.insertBefore(picker,addStyle);
    //Adds a break to put a new line after the color picker
    controls.insertBefore(document.createElement("br"),addStyle);
}

//Creates the tile grid
initGrid(ctx);

//Grabs the "save" button
const saveFace = document.getElementById("save");

/*
When "save" is clicked, run the save face sequence
 */
saveFace.onclick = () => {
    const face = {};
    //Set the face object's 'colors' field to the list of colors
    face['colors'] = styles;
    //Set the face object's 'pixels' field to the whole set of tiles
    face['pixels'] = pixels;
    //Download this face object
    downloadString(JSON.stringify(face));
}

//Fill each value in pixels with -1, the default color (white)
for (let i = 0 ; i < height ; i++ ){
    pixels[i] = new Array(width).fill(-1);
}

/*
Lets the user control via keyboard with arrow keys
 */
const handleMovement = (key) => {
    //Undoes highlighting of previously selected box
    highlightBox({x:x,y:y},true);
    //Checks what the key pressed was
    switch (key){
        //if right arrow, moves x to the right, same follows for others
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
    //If x is too far over, wraps back around to the other side
    if (x >= width){
        x = 0;
    }else if (x < 0){
        x = width - 1;
    }
    //If y is too far over, wraps back around to the other side
    if (y >= height){
        y = 0;
    }else if (y < 0){
        y = height - 1;
    }

    //Highlights the newly selected box
    highlightBox({x:x,y:y},false);
    //If color on toggle mode is true, draw the selected color wherever you moved to
    if (toggled){
        //Draw the current color here
        drawRect(ctx,{x:x,y:y},false);
    }

}

/*
When the user presses any key while on the page
 */
window.addEventListener("keydown",(event) => {
    //Get the keystroke of what they pressed
    const key = event.key;
    //If it starts with "Arrow"
    if (key.substr(0,5) === "Arrow"){
        //Move the cursor
        handleMovement(key);
    }else if (key === " "){
        //If it was a space bar, draw the rectangle at the current location
        drawRect(ctx,{x:x,y:y},true);
    }
})

//Highlight the default box (0,0)
highlightBox({x:x,y:y},false);

//The checkmark to do toggling
const check = document.getElementById("draw-on-move");
/*
If anyone changes the checkmark, flip toggle mode
 */
check.onchange = () => {
    toggled = !toggled;
}