package main

import "errors"

//var PageSize := 64

type Page struct {
	items    []Item
	nItem    int
	prevPage *Page
	nextPage *Page
}

func (p *Page) SetNextPage(next *Page) {
	p.nextPage = next
}

func (p *Page) SetPrevPage(prev *Page) {
	p.prevPage = prev
}

func NewPage(items []Items, pageSize int) (page *Page, err error) {
	if len(items) > pageSize {
		return &Page{items: items[:pageSize], pageSize}, errors.New("tried to fit too many items in page")
	}

	return &Page{items: items, len(items)}, nil
}

func (p *Page) Get(key string) (string, bool) {
	return sliceGet(p.items, key)
}

func (p *Page) Put(key, value string) bool { //, pageSize int) (midKey string, secondPage *Page, inserted bool) {
	inserted = slicePut(&p.items, key, value)
	//p.items = p.items[:]
	if inserted {
		p.nItem++
		/*
		   if p.nItem == pageSize {
		     //TODO:split
		     mid := pageSize / 2
		     halfLen2 := pageSize - mid
		     newPage := Page(make([]Item,halfLen2),halfLen2,p,p,nextPage}
		     copy(newPage.Items,p.items[mid:])
		     //firstPageItems = make([]Item,mid)
		     //copy(firstPageItems,p.items[:mid])
		     p.items = p.items[:mid]//firstPageItems
		     p.nItem = mid
		     p.nextPage = &newPage
		     return newPage.items[0].Key, &newPage, inserted
		   }
		*/
	}
	return inserted //nil, nil, inserted
}

func (p *Page) Full(fanout int) bool {
	return p.nItem == fanout
}

func (p *Page) Split(fanout int) (midKey string, newPage *Page) {
	mid := fanout / 2
	halfLen2 := fanout - mid
	newPage := Page{make([]Item, halfLen2), halfLen2, p, p.nextPage}
	copy(newPage.Items, p.items[mid:])
	//firstPageItems = make([]Item,mid)
	//copy(firstPageItems,p.items[:mid])
	p.items = p.items[:mid] //firstPageItems
	p.nItem = mid
	p.nextPage = &newPage
	return newPage.items[0].Key, &newPage
}

func (p *Page) Delete(key string) bool {
	deleted := sliceDelete(&p.items, key)
	if deleted {
		p.nItem--
	}
	return deleted
}

func (p *Page) GetFirstPage(key string) *Page {
  //TODO: if first is equal to the key than return this page
  //if not, then check for first GE in previous page
  //if first GE not in previous page then return this, else return previous
}

func (p *Page)HasFirstGE(key string) (int, bool) {
  i := sliceFirstGE(p.items, key)
  if i == len(p.items) {
    return false
  }
  return true
}

//TODO: return an iterator for the a start end key
//



/*
func (p *Page)Insert(key, value string, pageSize int) (midKey string, secondPage *Page) {
  i := sliceFirstGE(p.items, key)
  if p.items[i].Key == key {
    p.items[i].Value = value
    return nil, nil
  } else if p.nItem + 1 == pageSize {
    newPage := newPage{make([]Item,pageSize),0,p,p.nextPage}
    mid := pageSize / 2
    if i < mid {
      copy(newPage.items,p.items[mid - 1:p.nItem])
      for j := mid - 1; j > i ; j-- {
        p.items[j] = p.items[j-1]
      }
      p.items[i] = Item{key,value}

    } else {
      n := copy(newPage.items,p.items[mid:i])
      copy(newPage.items[n+1:],p.items[i:])
      newPage.items[n] = Item{key,value}
    }
    newPage.nItem = pageSize - mid
    p.nItem = mid
    p.nextPage = &newPage
    return newPage.items[0].Key, &newPage
  } else {
    for j := p.nItem ; j > i ; j-- {
      p.items[j] = p.items[j-1]
    }
    p.items[i] = Item{key,value}
    p.nItem++
    return nil, nil
  }
}
*/
