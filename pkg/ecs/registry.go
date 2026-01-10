package ecs

import (
	"reflect"
)

type Registry struct {
	lastEntity Entity
	freeList   []Entity
	masks      map[Entity]Bitmask
	storages   map[ComponentID]any
	typeIDs    map[reflect.Type]ComponentID
	deleters   map[ComponentID]func(Entity)
}

func newRegistry() *Registry {
	return &Registry{
		masks:    make(map[Entity]Bitmask),
		storages: make(map[ComponentID]any),
		typeIDs:  make(map[reflect.Type]ComponentID),
		deleters: make(map[ComponentID]func(Entity)),
	}
}

func (r *Registry) createEntity() Entity {
	var e Entity
	if len(r.freeList) > 0 {
		e = r.freeList[len(r.freeList)-1]
		r.freeList = r.freeList[:len(r.freeList)-1]
	} else {
		r.lastEntity++
		e = r.lastEntity
	}
	r.masks[e] = Bitmask{}
	return e
}

func (r *Registry) removeEntity(e Entity) {
	mask, ok := r.masks[e]
	if !ok {
		return
	}

	mask.ForEachSet(func(id ComponentID) {
		if deleteFn, exists := r.deleters[id]; exists {
			deleteFn(e)
		}
	})

	delete(r.masks, e)
	r.freeList = append(r.freeList, e)
}

func assign[T any](r *Registry, e Entity, component T) {
	id := registerComponent[T](r)
	assignByID(r, e, id, component)
}

func assignByID[T any](r *Registry, e Entity, id ComponentID, component T) {
	r.masks[e] = r.masks[e].Set(id)
	storage := r.storages[id].(map[Entity]*T)
	c := component
	storage[e] = &c
}

func unassign[T any](r *Registry, e Entity) {
	var dummy T
	t := reflect.TypeOf(dummy)

	id, ok := r.typeIDs[t]
	if !ok {
		return
	}

	unassignByID[T](r, e, id)
}

func unassignByID[T any](r *Registry, e Entity, id ComponentID) {
	if storage, ok := r.storages[id].(map[Entity]*T); ok {
		delete(storage, e)
	}

	if mask, ok := r.masks[e]; ok {
		r.masks[e] = mask.Clear(id)
	}
}

func get[T any](r *Registry, e Entity) *T {
	var dummy T
	t := reflect.TypeOf(dummy)

	id, ok := r.typeIDs[t]
	if !ok {
		return nil
	}

	if mask, exists := r.masks[e]; !exists || !mask.IsSet(id) {
		return nil
	}

	storage := r.storages[id].(map[Entity]*T)
	return storage[e]
}

func registerComponent[T any](r *Registry) ComponentID {
	var dummy T
	t := reflect.TypeOf(dummy)
	if id, ok := r.typeIDs[t]; ok {
		return id
	}

	id := ComponentID(len(r.typeIDs))
	r.typeIDs[t] = id

	storage := make(map[Entity]*T)
	r.storages[id] = storage

	r.deleters[id] = func(e Entity) {
		delete(storage, e)
	}

	return id
}
