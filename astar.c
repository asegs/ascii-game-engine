#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#define MOVES 4



struct Node {
	struct Node * pathParent;
	int arrival;
	int remaining;
	int row;
	int col;
	struct Node * left;
	struct Node * right;
	struct Node * graphParent;
};

struct PriorityQueue {
	struct Node * head;
};


struct Coord {
        int row;
        int col;
};

struct Tile {
	int type;
	int visited;
	struct Coord * pos;
};

struct NewMaze{
	struct Tile ** maze;
	struct Tile * start;
	struct Tile * end;
};


struct Node * NodeNew(struct Node * pathParent,int arrival,int remaining,int row,int col){
	struct Node * n = malloc(sizeof(struct Node ));
	n->pathParent = pathParent;
	n->arrival = arrival;
	n->remaining = remaining;
	n->row = row;
	n->col = col;
	n->left = NULL;
	n->right = NULL;
	n->graphParent = NULL;
	return n;
}

struct PriorityQueue * PQueueNew(struct Node * head){
	struct PriorityQueue * pq = malloc(sizeof(struct PriorityQueue));
	pq->head = head;
	return pq;
}

struct Coord * CoordNew(int row, int col){
        struct Coord * c = malloc(sizeof(struct Coord));
        c->row = row;
        c->col = col;
        return c;
}


struct Tile * TileNew(int type,int visited,int row,int col){
	struct Tile * t = malloc(sizeof(struct Tile));
	t->type = type;
	t->visited = visited;
	t->pos = CoordNew(row,col);
	return t;
}

struct NewMaze * NewMazeNew(struct Tile ** maze,struct Tile * start,struct Tile * end){
	struct NewMaze * n = malloc(sizeof(struct NewMaze));
	n->maze = maze;
	n->start = start;
	n->end = end;
	return n;
}


int estimate(struct Node * n){
	return n->arrival + n->remaining;
}

int matches(struct Node * n1,struct Node * n2){
	return n1->row == n2->row && n1->col == n2->col;
}

void insertNodeHelper(struct Node * parent,struct Node * toInsert){
	if (matches(parent,toInsert)){
        printf("HIT.\n");
		if (estimate(toInsert) < estimate(parent)){
			parent->pathParent = toInsert->pathParent;
			parent->arrival = toInsert->arrival;
		}
        return;
	}
	if (estimate(toInsert) >= estimate(parent)){
		if (!parent->right){
			toInsert->graphParent = parent;
			parent->right = toInsert;
			return;
		}
		insertNodeHelper(parent->right,toInsert);
	}else {
		if (!parent->left){
			toInsert->graphParent = parent;
			parent->left = toInsert;
			return;
		}
		insertNodeHelper(parent->left,toInsert);
	}
}


void insert(struct PriorityQueue * queue,struct Node * node){
	if (!queue->head){
		queue->head = node;
	}else{
		insertNodeHelper(queue->head,node);
	}
}

struct Node * takeClosestHelper(struct Node * parent){
	if (!parent->left){
		if (parent->right){
			struct Node * parentRoot = parent->graphParent;
			parentRoot->left = parent->right;
			parent->right->graphParent = parentRoot;
		}else{
			parent->graphParent->left = NULL;
		}

		return parent;
	}else{
		return takeClosestHelper(parent->left);
	}
}

struct Node * pop(struct PriorityQueue * queue){
	if (!queue->head){
		return NULL;
	}
	if (!queue->head->left){
		struct Node * toReturn = queue->head;
		if (queue->head->right){
			queue->head = queue->head->right;
		}else{
			queue->head = NULL;
		}
		return toReturn;
	}
	return takeClosestHelper(queue->head);
}


struct Tile * mazeAccess(struct Tile ** maze,int row,int col,int cols){
	struct Tile * t =  maze[row * cols + col];
    return t;
}

double rndm(){
	return ((double) rand() / (RAND_MAX));
}


int randRange(int min,int max){
	return rand() % (max + 1 - min) + min;
}

struct NewMaze * generateMaze(int width,int height,double freq){
	struct Tile ** maze = malloc ((width * height) * sizeof (struct Tile *));
	for (int i = 0;i<(width*height);i++){
		int row = i / width;
		int col = i % width;
		struct Tile * t = (rndm()<freq) ? TileNew(1,1,row,col) : TileNew(0,0,row,col);
		maze[i] = t;
	}
	struct Tile * startTile = mazeAccess(maze,randRange(0,height-1),randRange(0,width-1),width);
	struct Tile * endTile = mazeAccess(maze,randRange(0,height-1),randRange(0,width-1),width);
	startTile->type = 2;
	startTile->visited = 1;
	endTile->type = 3;
	return NewMazeNew(maze,startTile,endTile);

}

int square(int n){
	return n*n;
}

int pythagDistance(int row1,int row2,int col1,int col2){
	return square(row2-row1) + square(col2-col1);
}

//issues here
int tileGood(struct Tile ** maze,struct Tile * tile,int width,int height){
	return tile && tile->pos && (0 <= tile->pos->row) && (tile->pos->row < height) && (0 <= tile->pos->col) && (tile->pos->col < width) && !(mazeAccess(maze,tile->pos->row,tile->pos->col,width)->visited);
}

struct Coord * getCoordsForPair(struct Coord * pos,int rowMod,int colMod){
	return CoordNew(pos->row + rowMod,pos->col + colMod);
}

//may need to free tiles
struct Tile ** getAdjacentValidTiles(struct Tile ** maze,int height,int width,int row, int col){
	static int mods [MOVES * 2] = {0,-1,-1,0,0,1,1,0};
	struct Tile ** adjacent = malloc(MOVES * sizeof(struct Tile *));
	for (int i = 0;i<MOVES * 2;i+=2){
		struct Tile * t = mazeAccess(maze,row+mods[i],col+mods[i+1],width);
		if (tileGood(maze,t,width,height)){
			adjacent[i/2] = t;
		}else{
			adjacent[i/2] = NULL;
		}

	}
	return adjacent;
}

struct Coord ** unwrapPath(struct Node * end){
	struct Coord ** path = malloc((end->arrival + 1) * sizeof(struct Coord *));
	while (end){
		path[end->arrival] = CoordNew(end->row,end->col);
		end = end->pathParent;
	}
	return path;
}


struct Coord ** astar(struct NewMaze * maze,int height,int width){
    clock_t begin = clock();
	struct Coord * startPos = maze -> start -> pos;
	struct Coord * endPos = maze -> end -> pos;
    printf("Start: %d,%d\n",startPos->row,startPos->col);
    printf("End: %d,%d\n",endPos->row,endPos->col);
	struct Node * startNode = NodeNew(NULL,0,pythagDistance(startPos->row,endPos->row,startPos->col,endPos->col),startPos->row,startPos->col);
    struct PriorityQueue * queue = PQueueNew(startNode);
	while (queue->head){
		struct Node * position = pop(queue);
		if (position->row == endPos->row && position->col == endPos->col){
            clock_t term = clock();
            printf("%ld\n",(term - begin));
            printf("%ld\n",CLOCKS_PER_SEC);
			return unwrapPath(position);
		}
		struct Tile * current = mazeAccess(maze->maze,position->row,position->col,width);
        printf("%d\n",current->visited);
		current->visited = 1;
		struct Tile ** adjacent = getAdjacentValidTiles(maze->maze,height,width,position->row,position->col);
		for (int i = 0;i<MOVES;i++){
			if (adjacent[i]){
				struct Node * adj = NodeNew(position,position->arrival+1,pythagDistance(adjacent[i]->pos->row,endPos->row,adjacent[i]->pos->col,endPos->col),adjacent[i]->pos->row,adjacent[i]->pos->col);
				insert(queue,adj);
			}
		}
	}
    printf("FAILED\n");
	return NULL;
}



int main(){
	struct NewMaze * maze = generateMaze(100,100,0.1);
	astar(maze,100,100);
}
	

