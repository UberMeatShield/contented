import { Component, OnInit, Input } from '@angular/core';
import { Container } from './container';
import { ContentedService } from './contented_service';
import { finalize } from 'rxjs/operators';
import { GlobalBroadcast } from './global_message';

// TODO: When styling out the search add a hover and hover text to make it
// more obvious when something can be clicked.
@Component({
  selector: 'admin-container-cmp',
  templateUrl: './admin_containers.ng.html',
})
export class AdminContainersCmp implements OnInit {
  public loading = false;
  public creatingTask = false;

  @Input() containers: Array<Container>;

  constructor(private service: ContentedService) {}

  ngOnInit() {
    this.loading = true;
    this.service
      .getContainers()
      .pipe(
        finalize(() => {
          this.loading = false;
        })
      )
      .subscribe({
        next: res => {
          this.containers = res.results;
        },
        error: err => {
          GlobalBroadcast.error('Failed to load containers', err);
        },
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
