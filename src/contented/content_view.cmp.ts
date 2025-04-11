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
  @Input() content: Content | undefined;
  @Input() forceWidth: number = 0;
  @Input() forceHeight: number = 0;
  @Input() visible: boolean = false;

  public maxWidth: number = 0;
  public maxHeight: number = 0;
  public loading: boolean = false;
  public error: string | null = null;
  constructor(
    public _service: ContentedService,
    public route: ActivatedRoute,
    public router: Router
  ) {}

  public ngOnInit() {
    this.route.paramMap.pipe().subscribe({
      next: (res: ParamMap) => {
        const contentID = parseInt(res.get('id') || '0', 10);
        if (contentID) {
          this.loadContent(contentID);
        }
      },
      error: err => {
        GlobalBroadcast.error('Failed to start the content view', err);
      },
    });
  }

  public loadContent(contentID: number) {
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
          this.error = `Could not find ${contentID}`;
        },
      });
  }
}
