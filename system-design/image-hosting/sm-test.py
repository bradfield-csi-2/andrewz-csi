from PIL import Image
import profile
#with Image.open("./rick_and_morty_rtj_copy.jpg") as im:
    #Pizigani_1367_Chart_10MB_large.jpg

def bench_test():
    n = 1000



#    with Image.open("Pizigani_1367_Chart_10MB_large.jpg") as im:
    with Image.open("./rick_and_morty_rtj_copy.jpg") as im:
        #im.rotate(45).show()
        # Provide the target width and height of the image
        (width, height) = (im.width * 2, im.height * 2)
        typ = Image.BICUBIC
        for i in range(n):
            resize(im, width, height, typ)

def resize(im, width, height, typ):
    im_resized = im.resize((width, height), typ)


#bench_test()
profile.run('bench_test()')
