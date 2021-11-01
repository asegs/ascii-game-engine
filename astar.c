#include <stdio.h>
#include <stdlib.h>
#include <time.h>

struct Node {
	Node * pathParent;
	int arrival;
	int remaining;
	int row;
	int col;
	Node * left;
	Node * right;
	Node * graphParent;
}

struct PriorityQueue {
	Node * head;
}

struct Tile {
	int type;
	int visited;
}

int estimate(Node * n){
	return n->arrival + n->remaining;
}

int matches(Node * n1,Node * n2){
	return n1->row == n2-> && n1->col == n2->col;
}

void insertNodeHelper(Node * parent,Node * toInsert){
	if (matches(parent,toInsert)){
		if (estimate(toInsert) < estimate(parent)){
			parent->pathParent = node->pathParent;
			parent->arrival = node->arrival;
		}
	}
	if (estimate(node) >= estimate(parent)){
		if (!parent->right){
			node->graphParent = parent;
			parent->right = node;
			return;
		}
		insertNodeHelper(parent->right,node);
	}else {
		if (!parent->left){
			node->graphParent = parent;
			parent->left = node;
			return;
		}
		insertNodeHelper(parent->left,node);
	}
}


void insert(PriorityQueue * queue,Node * node){
	if (!node->head){
		queue->head = node;
	}else{
		insertNodeHelper(queue->head,node);
	}
}

Node * takeClosestHelper(Node * parent){
	if (!parent->left){
		if (parent->right){
			Node * parentRoot = parent->graphParent;
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

Node * pop(PriorityQueue * queue){
	if (!queue->head){
		return NULL;
	}
	if (!queue->head->left){
		Node * toReturn = queue->head;
		if (queue->head->right){
			queue->head = queue->head->right;
		}else{
			queue->head = NULL;
		}
		return toReturn;
	}
	return takeClosestHelper(queue->head);
}



Tile ** generateMaze(int width,int height,float freq){
	Tile ** maze = malloc ((width * height) * sizeof (Tile *));
	
}
