import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { DebugElement } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

import { ErrorHandlerCmp } from '../contented/error_handler.cmp';
import { ContentedService } from '../contented/contented_service';
import { ContentedModule } from '../contented/contented_module';

import _ from 'lodash';
import $ from 'jquery';
import { GlobalBroadcast } from './global_message';

describe('TestingErrorHandlerCmp', () => {
  let fixture: ComponentFixture<ErrorHandlerCmp>;
  let comp: ErrorHandlerCmp;
  let el: HTMLElement;
  let de: DebugElement;

  let httpMock: HttpTestingController;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      imports: [FormsModule, ContentedModule, HttpClientTestingModule, NoopAnimationsModule],
      providers: [ContentedService],
    }).compileComponents();

    fixture = TestBed.createComponent(ErrorHandlerCmp);
    httpMock = TestBed.inject(HttpTestingController);
    comp = fixture.componentInstance;

    de = fixture.debugElement.query(By.css('.error-handler-cmp'));
    el = de.nativeElement;
  }));

  afterEach(() => {
    httpMock.verify();
  });

  it('Should create a contented component', () => {
    expect(comp).withContext('We should have the error handler comp').toBeDefined();
    expect(el).withContext('We should have a top level element').toBeDefined();
  });

  it('Can we render errors and prevent duplicates', fakeAsync(() => {
    fixture.detectChanges();
    GlobalBroadcast.error('Boom', { thing: 'bad' });
    GlobalBroadcast.error('Boom', { thing: 'bad' });
    GlobalBroadcast.error('Satellite', { thing: 'bad' });
    GlobalBroadcast.evt('Not Boom', { thing: 'bad' });

    fixture.detectChanges();
    tick(1000);
    tick(1000);

    expect($('.error-count').length).withContext('The button should have a count').toEqual(1);
    expect($('.error-count').text()).withContext('We should have some errors').toEqual('2');
  }));
});
