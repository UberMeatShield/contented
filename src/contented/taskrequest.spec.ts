import { ComponentFixture, TestBed } from '@angular/core/testing';
import {NoopAnimationsModule} from '@angular/platform-browser/animations';
import {HttpClientTestingModule, HttpTestingController} from '@angular/common/http/testing';
import {By} from '@angular/platform-browser';
import {DebugElement} from '@angular/core';

import {ContentedModule} from '../contented/contented_module';
import { TaskRequestCmp } from './taskrequest.cmp';

describe('TaskRequestCmp', () => {
  let component: TaskRequestCmp;
  let fixture: ComponentFixture<TaskRequestCmp>;

  let httpMock: HttpTestingController;
  let el: HTMLElement;
  let de: DebugElement;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        NoopAnimationsModule,
        ContentedModule,
        HttpClientTestingModule,
      ],
      declarations: [TaskRequestCmp]
    });
    fixture = TestBed.createComponent(TaskRequestCmp);
    component = fixture.componentInstance;
    fixture.detectChanges();

    httpMock = TestBed.inject(HttpTestingController);
    de = fixture.debugElement.query(By.css('.task-request-cmp'));
    el = de.nativeElement;
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
