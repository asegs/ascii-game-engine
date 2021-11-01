#include <stdio.h>
#include <stdlib.h>
#include <time.h>



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

struct Tile {
	int type;
	int visited;
};

struct Node * NodeNew(struct Node * pathParent,int arrival,int remaining,int row,int col,struct Node * left,struct Node * right,struct Node * graphParent){
	struct Node * n = malloc(sizeof(struct Node ));
	n->pathParent = pathParent;
	n->arrival = arrival;
	n->remaining = remaining;
	n->row = row;
	n->col = col;
	n->left = left;
	n->right;
	n->graphParent = graphParent;
	return n;
}

struct PriorityQueue * PQueueNew(struct Node * head){
	struct PriorityQueue * pq = malloc(sizeof(struct PriorityQueue));
	pq->head = head;
	return pq;
}

struct Tile * TileNew(int type,int visited){
	struct Tile * t = malloc(sizeof(struct Tile));
	t->type = type;
	t->visited = visited;
	return t;
}

int estimate(struct Node * n){
	return n->arrival + n->remaining;
}

int matches(struct Node * n1,struct Node * n2){
	return n1->row == n2->row && n1->col == n2->col;
}

void insertNodeHelper(struct Node * parent,struct Node * toInsert){
	if (matches(parent,toInsert)){
		if (estimate(toInsert) < estimate(parent)){
			parent->pathParent = toInsert->pathParent;
			parent->arrival = toInsert->arrival;
		}
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


struct Tile * mazeAccess(struct Tile * maze,int row,int col,int cols){
	return &maze[row * cols + col];
}

double rndm(){
	return ((double) rand() / (RAND_MAX));
}


struct Tile * generateMaze(int width,int height,double freq){
	struct Tile * maze = malloc ((width * height) * sizeof (int));
	for (int i = 0;i<(width*height);i++){
		struct Tile * t = TileNew(0,0);
	}
	return maze;

}

int main(){
	;
}
	

