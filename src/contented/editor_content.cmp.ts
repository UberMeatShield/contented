/**
 * Provide the ability to edit the descriptions of content and containers.  Also provide the ability
 * to quickly manage tags.  TODO: This component should actually be broken into a pure wrapper around
 * the ngx-monaco intialization and handle just readonly and change emitting.
 */
import { Component, OnInit, AfterViewInit, Input, Output, EventEmitter, ViewChild} from '@angular/core';
import {ActivatedRoute, Router, ParamMap} from '@angular/router';
import {FormBuilder, FormControl, FormGroup, Validators} from '@angular/forms';
import {finalize, debounceTime, distinctUntilChanged} from 'rxjs/operators';
import {ContentedService} from './contented_service';
import {Content} from './content';

import * as _ from 'lodash-es';

@Component({
  selector: 'editor-content-cmp',
  templateUrl: './editor_content.ng.html',
})
export class EditorContentCmp implements OnInit {

  @Input() content?: Content;
  @Input() editForm?: FormGroup;
  @Input() descriptionControl: FormControl = new FormControl("", Validators.required);

  // These are values for the Monaco Editors, change events are passed down into
  // the form event via the AfterInit and set the v7_definition & suricata_definition.
  public loading: boolean = false;


  constructor(public fb: FormBuilder, public route: ActivatedRoute, public _service: ContentedService) {
    this.editForm = this.fb.group({
      description: this.descriptionControl
    });
  }

  // Subscribe to options changes, if the definition changes make the call
  public ngOnInit() {
    if (!this.content) {
        this.route.paramMap.pipe().subscribe(
            (map: ParamMap) => {
                this.loadContent(map.get('id'));
            },
            console.error
        );
    }
  }

  loadContent(id: string) {
      this._service.getContent(id).subscribe(
          (content: Content) => {
              this.content = content;
              this.descriptionControl.setValue(content.description);
          },
          console.error
      )
  }

  save() {
    console.log("Save()", this.editForm.value);
    this.content.description = _.get(this.editForm.value, 'description');
    this.loading = true;
    this._service.saveContent(this.content).pipe(finalize(() => this.loading = false)).subscribe(
      console.log,
      console.error
    );
  }
}
