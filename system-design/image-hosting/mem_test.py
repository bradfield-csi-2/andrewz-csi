from PIL import Image
#import profile
#with Image.open("./rick_and_morty_rtj_copy.jpg") as im:
    #Pizigani_1367_Chart_10MB_large.jpg

@profile
def bench_test():
    #n = 1000


    with Image.open("800px-Pizigani_1367_Chart_10MB_small.jpg") as im:
    #with Image.open("Pizigani_1367_Chart_10MB_large.jpg") as im:
    #with Image.open("Pizigani_1367_Chart_10MB_huge.jpg") as im:
    #with Image.open("./rick_and_morty_rtj_copy.jpg") as im:
        #im.rotate(45).show()
        # Provide the target width and height of the image
        target = 1024
        scale = target / im.width
        
        (width, height) = (target, int(im.height*scale) + 1 )
        typ = Image.BICUBIC
        #for i in range(n):
        resize(im, width, height, typ)

def resize(im, width, height, typ):
    im_resized = im.resize((width, height), typ)

if __name__ == '__main__':
    bench_test()
#bench_test()
#profile.run('bench_test()')

