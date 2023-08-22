import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import {By} from '@angular/platform-browser';
import {HttpClientTestingModule, HttpTestingController} from '@angular/common/http/testing';
import {DebugElement} from '@angular/core';

import { RouterTestingModule } from '@angular/router/testing';
import {ContainerNavCmp} from '../contented/container_nav_cmp';
import {Container} from '../contented/container';

import {ApiDef} from '../contented/api_def';
import {ContentedService} from '../contented/contented_service';
import {ContentedModule} from '../contented/contented_module';

import * as _ from 'lodash';
import * as $ from 'jquery';
import {MockData} from '../test/mock/mock_data';

describe('TestingContainerNavCmp', () => {
    let fixture: ComponentFixture<ContainerNavCmp>;
    let service: ContentedService;
    let comp: ContainerNavCmp;
    let el: HTMLElement;
    let de: DebugElement;
    let httpMock: HttpTestingController;

    beforeEach(waitForAsync( () => {
        TestBed.configureTestingModule({
            imports: [RouterTestingModule, ContentedModule, HttpClientTestingModule],
            providers: [
                ContentedService
            ],
            teardown: {destroyAfterEach: false},
        }).compileComponents();

        service = TestBed.inject(ContentedService);
        httpMock = TestBed.inject(HttpTestingController);
        fixture = TestBed.createComponent(ContainerNavCmp);
        comp = fixture.componentInstance;

        de = fixture.debugElement.query(By.css('.container-nav-cmp'));
        el = de.nativeElement;
    }));

    it('Should create a contented component', () => {
        expect(comp).toBeDefined("We should have the Contented comp");
        expect(el).toBeDefined("We should have a top level element");
    });

    it('Should be able to load up the basic data and render an image for each', fakeAsync(() => {
        let cnt = new Container(MockData.getPreview()[0]);
        cnt.total = 10;
        comp.cnt = cnt;

        fixture.detectChanges();
        expect($('.container-bar').length).toEqual(1, "It should have a container bar"); 
        let fullLoadBtn = $('.btn-full-load-ctn');
        expect(fullLoadBtn.length).toEqual(1, "We should have a full load btn");
        fullLoadBtn.trigger('click');
        fixture.detectChanges();

        let url = ApiDef.contented.content.replace("{cId}", cnt.id);
        let req = httpMock.expectOne(r => r.url === url);
    }));
});

