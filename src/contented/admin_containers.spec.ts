import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { DebugElement } from '@angular/core';

import { RouterTestingModule } from '@angular/router/testing';
import { AdminContainersCmp } from '../contented/admin_containers.cmp';
import { Container } from '../contented/container';

import { ContentedService } from '../contented/contented_service';
import { ContentedModule } from '../contented/contented_module';

import * as _ from 'lodash';
import * as $ from 'jquery';
import { MockData } from '../test/mock/mock_data';

describe('TestingAdminAdminContainersCmp', () => {
  let fixture: ComponentFixture<AdminContainersCmp>;
  let service: ContentedService;
  let comp: AdminContainersCmp;
  let el: HTMLElement;
  let de: DebugElement;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      imports: [RouterTestingModule, ContentedModule, HttpClientTestingModule],
      providers: [ContentedService],
      teardown: { destroyAfterEach: false },
    }).compileComponents();

    service = TestBed.get(ContentedService);
    fixture = TestBed.createComponent(AdminContainersCmp);
    comp = fixture.componentInstance;

    de = fixture.debugElement.query(By.css('.admin-containers-cmp'));
    el = de.nativeElement;
  }));

  it('Should create a contented component', () => {
    expect(comp).toBeDefined('We should have the Contented comp');
    expect(el).toBeDefined('We should have a top level element');

  });
});
