import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { DebugElement } from '@angular/core';

import { RouterTestingModule } from '@angular/router/testing';
import { AdminContainersCmp } from '../contented/admin_containers.cmp';
import { Container } from '../contented/container';

import { ContentedService } from '../contented/contented_service';
import { ContentedModule } from '../contented/contented_module';

import * as _ from 'lodash';
import * as $ from 'jquery';
import { MockData } from '../test/mock/mock_data';
import { ApiDef } from './api_def';

describe('TestingAdminAdminContainersCmp', () => {
  let fixture: ComponentFixture<AdminContainersCmp>;
  let service: ContentedService;
  let comp: AdminContainersCmp;
  let el: HTMLElement;
  let de: DebugElement;
  let httpMock: HttpTestingController;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      imports: [RouterTestingModule, ContentedModule, HttpClientTestingModule],
      providers: [ContentedService],
      teardown: { destroyAfterEach: false },
    }).compileComponents();

    service = TestBed.get(ContentedService);
    fixture = TestBed.createComponent(AdminContainersCmp);

    httpMock = TestBed.get(HttpTestingController);
    comp = fixture.componentInstance;

    de = fixture.debugElement.query(By.css('.admin-containers-cmp'));
    el = de.nativeElement;
  }));

  afterEach(() => {
    httpMock.verify();
  });

  it('Should create a contented component', () => {
    expect(comp).withContext('We should have the Contented comp').toBeDefined();
    expect(el).withContext('We should have a top level element').toBeDefined();
  });

  it('It can load up some containers', () => {
    fixture.detectChanges();

    const containers = MockData.getContainers();
    const req = httpMock.expectOne(ApiDef.contented.containers);
    req.flush(containers);
    expect(containers?.results?.length).toBeGreaterThan(0);

    fixture.detectChanges();
    expect($('.admin-cnt').length).toEqual(containers.results.length);
  });
});
