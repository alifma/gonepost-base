'use client';

import { useState } from 'react';

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle
} from '@/components/ui/alert-dialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Textarea } from '@/components/ui/textarea';
import { useDebounce } from '@/hooks/use-debounce';
import { useCan } from '@/lib/permissions';
import { useCreateItem, useDeleteItem, useItems, useUpdateItem, type Item, type ItemStatus } from '../api';

const PAGE_SIZE = 10;

export default function ItemsPage() {
  const { can } = useCan();
  const canWrite = can('items:write');

  const [search, setSearch] = useState('');
  const [status, setStatus] = useState<ItemStatus | ''>('');
  const [page, setPage] = useState(1);
  const debouncedSearch = useDebounce(search, 300);

  const items = useItems({
    search: debouncedSearch || undefined,
    status: status || undefined,
    limit: PAGE_SIZE,
    offset: (page - 1) * PAGE_SIZE
  });
  const total = items.data?.total ?? 0;
  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE));

  // `null` = dialog closed, `'new'` = create, an Item = edit that item
  const [editing, setEditing] = useState<Item | 'new' | null>(null);
  const [deleting, setDeleting] = useState<Item | null>(null);
  const remove = useDeleteItem();

  return (
    <>
      <section className='border-border/70 bg-card overflow-hidden rounded-xl border'>
        <div className='border-border/70 flex flex-wrap items-center gap-2 border-b px-5 py-4'>
          <Input
            className='max-w-xs'
            placeholder='Search items...'
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
          />
          <NativeSelect
            value={status}
            onChange={(e) => {
              setStatus(e.target.value as ItemStatus | '');
              setPage(1);
            }}
          >
            <NativeSelectOption value=''>All statuses</NativeSelectOption>
            <NativeSelectOption value='active'>Active</NativeSelectOption>
            <NativeSelectOption value='archived'>Archived</NativeSelectOption>
          </NativeSelect>
          {canWrite && (
            <Button size='sm' className='ml-auto' onClick={() => setEditing('new')}>
              New item
            </Button>
          )}
        </div>

        {items.isError && (
          <p className='text-destructive px-5 pt-4 text-sm'>
            Items could not be loaded: {items.error.message}
          </p>
        )}

        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Status</TableHead>
              {canWrite && <TableHead className='text-right'>Actions</TableHead>}
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.data?.data.map((item) => (
              <TableRow key={item.id}>
                <TableCell className='font-medium'>{item.name}</TableCell>
                <TableCell className='text-muted-foreground max-w-md truncate'>
                  {item.description ?? '-'}
                </TableCell>
                <TableCell>
                  <Badge variant={item.status === 'active' ? 'default' : 'secondary'}>{item.status}</Badge>
                </TableCell>
                {canWrite && (
                  <TableCell className='space-x-2 text-right'>
                    <Button variant='outline' size='sm' onClick={() => setEditing(item)}>
                      Edit
                    </Button>
                    <Button variant='outline' size='sm' onClick={() => setDeleting(item)}>
                      Delete
                    </Button>
                  </TableCell>
                )}
              </TableRow>
            ))}
            {items.isSuccess && items.data.data.length === 0 && (
              <TableRow>
                <TableCell colSpan={canWrite ? 4 : 3} className='text-muted-foreground py-8 text-center'>
                  No items found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>

        <div className='border-border/70 flex items-center justify-between border-t px-5 py-3 text-sm'>
          <span className='text-muted-foreground'>
            {total} item{total === 1 ? '' : 's'}
          </span>
          <div className='flex items-center gap-2'>
            <Button variant='outline' size='sm' disabled={page <= 1} onClick={() => setPage(page - 1)}>
              Previous
            </Button>
            <span>
              {page} / {pageCount}
            </span>
            <Button variant='outline' size='sm' disabled={page >= pageCount} onClick={() => setPage(page + 1)}>
              Next
            </Button>
          </div>
        </div>
      </section>

      {editing && <ItemDialog item={editing === 'new' ? null : editing} onClose={() => setEditing(null)} />}

      <AlertDialog open={!!deleting} onOpenChange={(open) => !open && setDeleting(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete {deleting?.name}?</AlertDialogTitle>
            <AlertDialogDescription>This cannot be undone.</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => deleting && remove.mutate(deleting.id, { onSettled: () => setDeleting(null) })}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}

/** Create (item = null) or edit form. Mounted only while open, so its state resets every time. */
function ItemDialog({ item, onClose }: { item: Item | null; onClose: () => void }) {
  const create = useCreateItem();
  const update = useUpdateItem();
  const pending = create.isPending || update.isPending;

  const [name, setName] = useState(item?.name ?? '');
  const [description, setDescription] = useState(item?.description ?? '');
  const [status, setStatus] = useState<ItemStatus>(item?.status ?? 'active');

  function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const body = { name, description, status };
    const done = { onSuccess: onClose };
    if (item) update.mutate({ id: item.id, body }, done);
    else create.mutate(body, done);
  }

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{item ? 'Edit item' : 'New item'}</DialogTitle>
          <DialogDescription>{item ? 'Change this item.' : 'Add an item to your list.'}</DialogDescription>
        </DialogHeader>
        <form className='grid gap-4' onSubmit={submit}>
          <div className='grid gap-2'>
            <Label htmlFor='item-name'>Name</Label>
            <Input id='item-name' value={name} onChange={(e) => setName(e.target.value)} required maxLength={120} />
          </div>
          <div className='grid gap-2'>
            <Label htmlFor='item-description'>Description</Label>
            <Textarea
              id='item-description'
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              maxLength={1000}
            />
          </div>
          <div className='grid gap-2'>
            <Label htmlFor='item-status'>Status</Label>
            <NativeSelect id='item-status' value={status} onChange={(e) => setStatus(e.target.value as ItemStatus)}>
              <NativeSelectOption value='active'>Active</NativeSelectOption>
              <NativeSelectOption value='archived'>Archived</NativeSelectOption>
            </NativeSelect>
          </div>
          <DialogFooter>
            <Button type='submit' disabled={pending}>
              {pending ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
