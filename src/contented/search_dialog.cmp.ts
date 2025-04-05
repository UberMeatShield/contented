import { Component, OnInit, AfterViewInit, ViewChild, Inject, ElementRef } from '@angular/core';
import { Content } from './content';
import { ContentedService } from './contented_service';
import { MAT_DIALOG_DATA } from '@angular/material/dialog';
import { initializeDefaults } from './utils';

@Component({
    selector: 'search-dialog',
    templateUrl: 'search_dialog.ng.html',
    standalone: false
})
export class SearchDialog implements AfterViewInit {
  public contentContainer: Content;
  public forceHeight: number = 0;
  public forceWidth: number = 0;
  public sizeCalculated: boolean = false;

  @ViewChild('SearchContent', { static: true }) searchContent: ElementRef | undefined;

  constructor(
    public _service: ContentedService,
    @Inject(MAT_DIALOG_DATA) public data: Content
  ) {
    this.contentContainer = data;
  }

  ngAfterViewInit() {
    // TODO: Sizing content is a little off and the toolbars are visible based on dialog size
    setTimeout(() => {
      const el = this.searchContent?.nativeElement;
      if (el) {
        console.log('Element', el, el.offsetWidth, el.offsetHeight);
        this.forceHeight = el.offsetHeight - 40;
        this.forceWidth = el.offsetWidth - 40;
      }
      this.sizeCalculated = true;
    }, 100);
  }
} 