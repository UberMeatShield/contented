import {async, fakeAsync, getTestBed, tick, ComponentFixture, TestBed} from '@angular/core/testing';
import {By} from '@angular/platform-browser';
import {HttpClientTestingModule} from '@angular/common/http/testing';
import {DebugElement} from '@angular/core';

import { RouterTestingModule } from '@angular/router/testing';
import {ContentedCmp} from '../contented/contented_cmp';
import {ContentedService} from '../contented/contented_service';
import {ContentedModule} from '../contented/contented_module';

import * as _ from 'lodash';
import {MockData} from './mock/mock_data';

declare var $;
describe('TestingContentedsCmp', () => {
    let fixture: ComponentFixture<ContentedCmp>;
    let service: ContentedService;
    let comp: ContentedCmp;
    let el: HTMLElement;
    let de: DebugElement;

    beforeEach(async( () => {
        TestBed.configureTestingModule({
            imports: [RouterTestingModule, ContentedModule, HttpClientTestingModule],
            providers: [
                ContentedService
            ]
        }).compileComponents();

        service = TestBed.get(ContentedService);
        fixture = TestBed.createComponent(ContentedCmp);
        comp = fixture.componentInstance;

        de = fixture.debugElement.query(By.css('.contented-cmp'));
        el = de.nativeElement;
    }));

    it('Should create a contented component', () => {
        expect(comp).toBeDefined("We should have the Contented comp");
        expect(el).toBeDefined("We should have a top level element");
    });

    it('Should be able to load up the basic data and render an image for each', fakeAsync(() => {
        MockData.mockContentedService(comp._contentedService);
        fixture.detectChanges();
        tick(2000);
        expect(comp.allD.length).toBe(3, "We should have 3 directories set");

        let dirs = comp.getVisibleDirectories();
        expect(dirs.length).toBe(comp.maxVisible, "Should only have the max visible directories present.");

        fixture.detectChanges();
        let dirEls = $('.directory-contents', el);
        expect(dirEls.length).toBe(2, "The UI should contain exactly 2 preview directories.");
    }));
});

