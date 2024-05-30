import { OnInit, Component, Input } from '@angular/core';
import { Content } from './content';
import { ContentedService } from './contented_service';
import { ActivatedRoute, Router, ParamMap } from '@angular/router';

import { finalize } from 'rxjs/operators';
import { GlobalBroadcast } from './global_message';

@Component({
  selector: 'content-view',
  templateUrl: './content_view.ng.html',
})
export class ContentViewCmp implements OnInit {
  @Input() content: Content;
  @Input() forceWidth: number;
  @Input() forceHeight: number;
  @Input() visible: boolean = false;

  public maxWidth: number;
  public maxHeight: number;
  public loading: boolean = false;
  public error = null;

  constructor(
    public _service: ContentedService,
    public route: ActivatedRoute,
    public router: Router
  ) {}

  public ngOnInit() {
    this.route.paramMap.pipe().subscribe({
      next: (res: ParamMap) => {
        let contentID = res.get('id');
        if (contentID) {
          this.loadContent(contentID);
        }
      },
      error: err => {
        GlobalBroadcast.error('Failed to start the content view', err);
      },
    });
  }

  public loadContent(contentID: string) {
    this.loading = true;
    this._service
      .getContent(contentID)
      .pipe(
        finalize(() => {
          this.loading = false;
        })
      )
      .subscribe({
        next: (m: Content) => {
          this.content = m;
        },
        error: err => {
          GlobalBroadcast.error(`Could not find ${contentID}`, err);
        },
      });
  }
}
