import { Component, OnInit, Input } from '@angular/core';
import { Container } from './container';
import { ContainerSearchSchema, ContentedService } from './contented_service';
import { finalize } from 'rxjs/operators';
import { GlobalBroadcast } from './global_message';
import { Tag, VSCodeChange } from './content';

import * as _ from 'lodash';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { PageEvent } from '@angular/material/paginator';
// TODO: When styling out the search add a hover and hover text to make it
// more obvious when something can be clicked.
@Component({
  selector: 'admin-container-cmp',
  templateUrl: './admin_containers.ng.html',
  standalone: false,
})
export class AdminContainersCmp implements OnInit {
  @Input() tags: Array<Tag> = [];
  @Input() containers: Array<Container> = [];

  public loading = false;
  public creatingTask = false;
  changedSearch: (evt: VSCodeChange) => void = () => {};
  currentTextChange: VSCodeChange = { value: '', tags: [] };

  options: FormGroup | undefined;
  searchType = new FormControl('text');
  searchText: string = '';

  public total = 0;
  public page = 1; // Tracking current page
  public pageSize = 50;

  constructor(
    private service: ContentedService,
    private fb: FormBuilder
  ) {
    this.resetForm();
  }

  ngOnInit() {
    this.loading = true;

    this.resetForm();
    this.changedSearch = _.debounce((evt: VSCodeChange) => {
      // Do not change this.searchText it will re-assign the VS-Code editor in a
      // bad way and muck with the cursor.
      if (this.currentTextChange.value != evt.value) {
        this.search(evt.value, this.page, this.pageSize, evt.tags);
      }
      this.currentTextChange = evt;
    }, 250);

    this.search(this.currentTextChange.value, this.page, this.pageSize, this.currentTextChange.tags);
  }

  search(search: string, page: number, perPage: number, tags: Array<string> = []) {
    console.log('SEARCH');
    const query = ContainerSearchSchema.parse({
      search,
      page,
      per_page: perPage,
      tags,
    });
    this.loading = true;
    this.service
      .searchContainers(query)
      .pipe(
        finalize(() => {
          this.loading = false;
        })
      )
      .subscribe({
        next: res => {
          console.log('Results should assign containers', res.results, 'total', res.total);
          this.containers = res.results;
        },
        error: err => {
          GlobalBroadcast.error('Failed to search containers', err);
        },
      });
  }

  pageEvt(evt: PageEvent) {
    console.log('Event', evt, this.currentTextChange.value);
    let page = evt.pageIndex + 1; // Angular Material uses 0-based index, we use 1-based
    let perPage = evt.pageSize;
    this.search(this.currentTextChange.value, page, perPage, this.currentTextChange.tags);
  }

  public resetForm(setupFilterEvents: boolean = false) {
    this.options = this.fb.group({
      searchType: this.searchType,
    });
  }

  createPreviews(cnt: Container) {
    this.creatingTask = true;
    this.service
      .containerPreviewsTask(cnt)
      .pipe(finalize(() => (this.creatingTask = false)))
      .subscribe({
        next: response => {
          console.log('Queued', response);
        },
        error: err => {
          GlobalBroadcast.error('Failed to start previews task', err);
        },
      });
  }

  createWebp(cnt: Container) {
    this.creatingTask = true;
    console.log('Create Webp');
    this.creatingTask = false;
  }

  createTags(cnt: Container) {
    this.creatingTask = true;
    this.service
      .containerTaggingTask(cnt)
      .pipe(finalize(() => (this.creatingTask = false)))
      .subscribe({
        next: response => {
          console.log('Queued', response);
        },
        error: err => {
          GlobalBroadcast.error('Failed to start tagging task', err);
        },
      });
  }

  // This one is awkward, kinda need to show this directly in the container.
  findDuplicates(cnt: Container) {
    this.creatingTask = true;
    this.service
      .containerDuplicatesTask(cnt)
      .pipe(finalize(() => (this.creatingTask = false)))
      .subscribe({
        next: response => {
          console.log('Queued', response);
        },
        error: err => {
          GlobalBroadcast.error('Failed to start duplicates task', err);
        },
      });
  }

  removeDuplicates(cnt: Container) {
    this.creatingTask = true;
    this.service
      .containerRemoveDuplicatesTask(cnt)
      .pipe(finalize(() => (this.creatingTask = false)))
      .subscribe({
        next: response => {
          console.log('Queued', response);
        },
        error: err => {
          GlobalBroadcast.error('Failed to start duplicates task', err);
        },
      });
  }

  encodeVideos(cnt: Container) {
    console.log('Encode Videos', cnt);
    this.creatingTask = true;
    this.service
      .containerVideoEncodingTask(cnt)
      .pipe(finalize(() => (this.creatingTask = false)))
      .subscribe({
        next: response => {
          console.log('Queued', response);
        },
        error: err => {
          GlobalBroadcast.error('Failed to start encoding tasks', err);
        },
      });
  }
}
