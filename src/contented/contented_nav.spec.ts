import { Subscription } from 'rxjs';
import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { DebugElement } from '@angular/core';

import { ContentedNavCmp } from '../contented/contented_nav.cmp';
import { Container } from '../contented/container';

import { ApiDef } from '../contented/api_def';
import { ContentedService } from '../contented/contented_service';
import { ContentedModule } from '../contented/contented_module';
import { GlobalNavEvents, NavTypes } from '../contented/nav_events';

import _ from 'lodash';
import $ from 'jquery';
import { MockData } from '../test/mock/mock_data';
import { provideRouter } from '@angular/router';
import { RouterTestingHarness } from '@angular/router/testing';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';

describe('TestingContentedNavCmp', () => {
  let fixture: ComponentFixture<ContentedNavCmp>;
  let service: ContentedService;
  let comp: ContentedNavCmp;
  let el: HTMLElement;
  let de: DebugElement;
  let httpMock: HttpTestingController;
  let sub: Subscription;

  beforeEach(waitForAsync(async () => {
    TestBed.configureTestingModule({
    teardown: { destroyAfterEach: true },
    imports: [ContentedModule],
    providers: [provideRouter([{ path: '', component: ContentedNavCmp }]), ContentedService, provideHttpClient(withInterceptorsFromDi()), provideHttpClientTesting()]
}).compileComponents();

    await RouterTestingHarness.create();

    service = TestBed.inject(ContentedService);
    httpMock = TestBed.inject(HttpTestingController);
    fixture = TestBed.createComponent(ContentedNavCmp);
    comp = fixture.componentInstance;

    de = fixture.debugElement.query(By.css('.contented-nav-cmp'));
    el = de.nativeElement;
  }));

  afterEach(() => {
    if (sub) {
      sub.unsubscribe();
      sub = null;
    }
  });

  it('Should create a contented component', () => {
    expect(comp).withContext('We should have the Contented comp').toBeDefined();
    expect(el).withContext('We should have a top level element').toBeDefined();
  });

  it('Should be able to handle certain key events', () => {
    let counter = 0;
    sub = GlobalNavEvents.navEvts.subscribe(evt => {
      console.log(evt);
      counter++;
    });

    comp.handleKey('w');
    comp.handleKey('a');
    comp.handleKey('s');
    comp.handleKey('d');

    comp.handleKey('e');
    comp.handleKey('q');
    comp.handleKey('x');
    comp.handleKey('f');
    comp.handleKey('t');
    fixture.detectChanges();

    fixture.whenStable().then(() => {
      expect(counter).withContext('It should have handled these events').toEqual(9);
    });
  });

  it('Should be able to handle a document keyup', fakeAsync(() => {
    fixture.detectChanges();

    let validate: NavTypes = null;
    sub = GlobalNavEvents.navEvts.subscribe(evt => {
      validate = evt.action;
    });

    document.dispatchEvent(new KeyboardEvent('keyup', { key: 'a' }));
    fixture.detectChanges();

    fixture.whenStable().then(() => {
      expect(validate).toEqual(NavTypes.PREV_MEDIA);
    });
    fixture.detectChanges();
    tick(2000);
    tick(2000);
  }));
});
